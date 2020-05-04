#!/usr/bin/python3

import asyncio
import aiohttp
from aiohttp.resolver import AsyncResolver
from bs4 import BeautifulSoup
from datetime import datetime, timedelta
from os.path import dirname, abspath, exists
from loguru import logger
from threading import Thread
from json import loads
from random import SystemRandom
from socket import AF_INET

from db import Database

headers = {
    'Referer' : 'google.com',
    'Host' : 'www.asu.ru',
    'Connection' : 'keep-alive',
    'Cache-Control' : 'max-age=0',
    'Upgrade-Insecure-Requests' : '1',
    'User-Agent' : 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36',
    'DNT' : '1',
    'Accept' : 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9',
    'Accept-Encoding' : 'gzip, deflate',
    'Accept-Language' : 'en-GB,en;q=0.9,ru-RU;q=0.8,ru;q=0.7,en-US;q=0.6'
    }

def start_background_loop(loop: asyncio.AbstractEventLoop):
    asyncio.set_event_loop(loop)
    loop.run_forever()          

class Loop:
    loop = asyncio.new_event_loop()
    main_Thread = Thread(target = start_background_loop, args = (loop,), daemon = True).start()  

class Page:
    @classmethod
    async def download(cls, link, session):
        try:    
            async with session.get(link, headers = headers) as request:
                if request.status == 200:
                    return await request.text()
                else:
                    return None

        except Exception as error:
            logger.exception(error)   

    @classmethod
    def find_Faculty_ID(cls, page, faculty) -> str:
        try:
            soup = BeautifulSoup(page, 'lxml')
            content = soup.find_all(class_ = 'padding_left_x')
        except TypeError:
            return None

        for line in content[0].find_all(class_ = 'link_ptr_left margin_bottom'):
            res = line.find('a', href = True)

            if faculty in res.text.lower():
                return res["href"].replace("/", "")

    @classmethod
    def find_Prep_ID(cls, page, prep) -> str:
        prep = prep.lower()

        try:
            soup = BeautifulSoup(page, 'lxml')
            content = soup.find_all(class_ = 'padding_left_x')
        except TypeError:
            return None

        for line in content:
            res = line.find_all(class_ = 'link_ptr_left margin_bottom')

            for line in res:
                search_Res = line.find('a', href = True)

                if prep in search_Res.text.lower().split(',')[0]:
                    return search_Res["href"]

    @classmethod
    def find_Cathedra_ID(cls, page, cathedra) -> str:
        cathedra = cathedra.lower()

        try:
            soup = BeautifulSoup(page, 'lxml')
            content = soup.find_all(class_ = 'padding_left_x')
        except TypeError:
            return None

        for line in content:
            res = line.find_all(class_ = 'link_ptr_left margin_bottom')

            for line in res:
                search_Res = line.find('a', href = True)

                if cathedra == search_Res.text.lower():
                    return search_Res["href"]

    # Поиск ссылки на расписание группы
    @classmethod                                                                                                                                                           
    def find_Group_ID(cls, page, group) -> str:     
        group = group.lower()

        try:
            soup = BeautifulSoup(page, 'lxml')
            content = soup.find_all(class_ = 'padding_left_x align_top')
        except TypeError:
            return None

        for line in content:
            res = line.find_all(class_ = 'link_ptr_left margin_bottom')

            for line in res:
                search_Res = line.find('a', href = True)

                if group == search_Res.text.lower():
                    return search_Res["href"]

    @classmethod
    async def download_Rasp(cls, link_For_Rasp, session):
        start_Next_Week = datetime.now() + timedelta(days = 7 - datetime.now().weekday())
        
        try:
            async with session.get(link_For_Rasp, headers = headers) as request:
                if request.status == 200:
                    this_Week = await request.text()
                else:
                    this_Week = None
                        
            async with session.get(
                f'{link_For_Rasp}?date={str(start_Next_Week).split(" ")[0].replace("-", "")}'
                f'-{str(start_Next_Week + timedelta(days = 6)).split(" ")[0].replace("-", "")}',
                headers = headers
            ) as request:
                if request.status == 200:
                    next_Week = await request.text()
                else:
                    next_Week = None

            return this_Week, next_Week

        except Exception as error:
            logger.exception(error)
             
    @classmethod
    def parse(cls, page, prep_Page):
        result = {}
        prev_Time = ''

        try:
            soup = BeautifulSoup(page, 'lxml')

            # На страничках преподов класс называется чуть иначе :)
            if prep_Page is True:
                page = soup.find_all(class_='schedule align_top')[0].find_all(class_='schedule-time')
            else:    
                page = soup.find_all(class_='align_top schedule')[0].find_all(class_='schedule-time')
        except IndexError:
            return -10
            
        for record in page:
            record = record.find_all('td')
            para_Cleaned = ''
            data = None
            
            # На страничках преподов нужно сдвинуть индекс для нахождения даты пары :)
            if prep_Page is True:
                tmp = record[6].find('a', href = True)
            else:
                tmp = record[5].find('a', href = True)
            
            if tmp is not None:
                # У групп 56* и 57* в ссылке почему-то ;building
                # а у групп 58* и 59* в ссылке почему-то &building
                tmp = tmp['href'].translate(str.maketrans({'&': '', ';': ''}))

                data = tmp.replace('/timetable/freerooms/?date=', '').replace('building', '')[0:8]
                data = str(datetime(int(data[0:4]), int(data[4:6]), int(data[6:8])).date())
            
            # Иногда на сайте не указана дата, просто пропускаем
            if data is None:
                continue

            auditoriya = record[4].text.replace('\n', '').rstrip()
            para_Name = record[2].text.replace('\n', '').split(' ')
            time = record[1].text.replace('\n', '').rstrip()

            # Если строка пуста (что бывает, когда у двух подгрупп пара в одно время)
            # то просто ставим предыдущее значение
            if time:
                prev_Time = time
            if not time:
                time = prev_Time

            for part in para_Name:
                if part != '':
                    para_Cleaned += part + ' '

            if data not in result:
                result[data] = {
                    'time' : [time],
                    'para': [para_Cleaned],
                    'auditoriya' : [auditoriya]
                }
                
            else:
                result[data]['time'].append(time)
                result[data]['para'].append(para_Cleaned)
                result[data]['auditoriya'].append(auditoriya)
                
        return result

class Raspisanie:
    # Страница с ссылками на все факультеты
    groups_Link = 'http://www.asu.ru/timetable/students/'
    prep_Link = 'http://www.asu.ru/timetable/lecturers/'
    cache = {}

    @classmethod
    async def init(cls, Path, database, refresh_Rate):
        tasks = []

        resolver = AsyncResolver(nameservers = ['8.8.8.8', '1.1.1.1'])
        generator = SystemRandom()

        tasks.append(asyncio.create_task(Raspisanie.Prepods.refresher(Path, database, refresh_Rate, resolver, generator)))
        tasks.append(asyncio.create_task(Raspisanie.Groups.refresher(Path, database, refresh_Rate, resolver, generator)))

        await asyncio.gather(*tasks)

    class Prepods:
        @classmethod
        async def refresher(cls, Path, database, refresh_Rate, dns_Resolver, generator):
            faculties = []
            cathedra = []
            preps = []

            try:
                if not exists(f'{Path}/data/preps'):
                    logger.critical('File with preps list not found')
                    exit(1)
                else:
                    with open(f'{Path}/data/preps', 'r') as file:
                        for line in file:
                            line = line.replace('\n', '').split(':')

                            faculties.append(line[0].lower())
                            cathedra.append(line[1].lower())
                            preps.append(line[2].lower())

                while True:
                    async with aiohttp.TCPConnector(
                        limit = None, ttl_dns_cache = 300, 
                        resolver = dns_Resolver, family = AF_INET, 
                        force_close = True
                    ) as connector:        
                        async with aiohttp.ClientSession(connector = connector) as session:
                            logger.info('started prep rasp refreshing')

                            tasks = [
                                asyncio.create_task(Raspisanie.Prepods.get(faculty, cathedra, prep, session, generator.uniform(0, 10)))
                                for faculty, cathedra, prep in zip(faculties, cathedra, preps)
                            ]

                            res = await asyncio.gather(*tasks)

                    for part in res:
                        if 'error' in part.keys():
                            logger.critical(f'Error - {part.get("error")}')
                        else:
                            await database.insert_Prep(part.get('prep'), part.get('payload'))

                    await database.fsync()

                    logger.info(f'sleeping {refresh_Rate}')
                    await asyncio.sleep(refresh_Rate)

            except Exception as error:
                logger.exception(error)

        @classmethod
        async def get(cls, faculty, cathedra, prep, session, delay):
            result = {}

            await asyncio.sleep(delay)

            logger.info(f'Getting prep {prep} rasp')

            try:
                # Находим ID факультета
                if Raspisanie.cache.get(faculty) is None:
                    faculty_Page = await Page.download(Raspisanie.prep_Link, session)

                    if faculty_Page == None:
                        return {'error': f'Не удалось скачать страницу с факультетами - {faculty}, {cathedra}, {prep}'}

                    faculty_ID = Page.find_Faculty_ID(faculty_Page, faculty)

                    if faculty_ID == None:
                        return {'error': f'Не удалось найти указанный факультет - {faculty}'}
                    else:
                        Raspisanie.cache[faculty] = faculty_ID

                else:
                    faculty_ID = Raspisanie.cache.get(faculty)

                # Ищем ID кафедры
                if Raspisanie.cache.get(cathedra) is None:
                    cathedra_Page = await Page.download(f'{Raspisanie.prep_Link}{faculty_ID}', session)

                    if cathedra_Page == None:
                        return {'error': f'Не удалось скачать страницу с кафедрами - {faculty}, {cathedra}, {prep}'}

                    cathedra_ID = Page.find_Cathedra_ID(cathedra_Page, cathedra)

                    if cathedra_ID == None:
                        return {'error': f'Не удалось найти указанную кафедру - {cathedra}'}
                    else:
                        Raspisanie.cache[cathedra] = cathedra_ID

                else:
                    cathedra_ID = Raspisanie.cache.get(cathedra)

                if Raspisanie.cache.get(prep) is None:
                    preps_Page = await Page.download(f'{Raspisanie.prep_Link}{faculty_ID}/{cathedra_ID}', session)

                    if preps_Page is None:
                        return {'error': f'Не удалось скачать страницу с преподавателями - {faculty}, {cathedra}, {prep}'}

                    prep_ID = Page.find_Prep_ID(preps_Page, prep)

                    if prep_ID == None:
                        return {'error': f'Не удалось найти указанного преподавателя - {prep}'}
                    else:
                        Raspisanie.cache[prep] = prep_ID
                else:
                    prep_ID = Raspisanie.cache.get(prep)

                link = f'{Raspisanie.prep_Link}{faculty_ID}/{cathedra_ID}{prep_ID}'

                try:
                    this_Week, next_Week = await Page.download_Rasp(link, session)
                except TypeError:
                    return {'error': f'Не нашел расписания для {prep}'}

                # Парсим полученные странички
                this_Week = Page.parse(this_Week, True)
                next_Week = Page.parse(next_Week, True)

                # Если парсер нашел нужное в страничках, то записываем их в переменные
                if this_Week != -10:
                    result.update(this_Week)
                if next_Week != -10:
                    result.update(next_Week)
                if this_Week == -10 and next_Week == -10:
                    return {'error': f'Не нашел расписания для {prep}'}

                logger.info(f'Rasp for prep {prep} is ready')

                return {
                    'prep' : prep,
                    'payload': result
                }

            except Exception as error:
                logger.exception(error)

    class Groups:
        @classmethod
        async def refresher(cls, Path, database, refresh_Rate, dns_Resolver, generator):
            groups = []
            faculties = []

            try:
                if not exists(f'{Path}/data/groups'):
                    logger.critical('File with groups list not found')
                    exit(1)
                else:
                    with open(f'{Path}/data/groups', 'r') as file:
                        for line in file:
                            line = line.replace('\n', '').split(':')

                            faculties.append(line[0].lower())
                            groups.append(line[1].lower())

                while True:
                    async with aiohttp.TCPConnector(
                        limit = None, ttl_dns_cache = 300, 
                        resolver = dns_Resolver, family = AF_INET, 
                        force_close = True
                    ) as connector:
                    
                        async with aiohttp.ClientSession(connector = connector) as session:
                            logger.info('started groups rasp refreshing')
                
                            tasks = [
                                asyncio.create_task(Raspisanie.Groups.get(group, faculty, session, generator.uniform(0, 10)))
                                for group, faculty in zip(groups, faculties)
                            ]

                            res = await asyncio.gather(*tasks)

                    for part in res:
                        if part is None:
                            logger.critical('Error - payload does\'nt contain any info')
                        elif 'error' in part.keys():
                            logger.critical(f'Error - {part.get("error")}')
                        else:
                            await database.insert(part.get('group'), part.get('payload'))

                    await database.fsync()

                    logger.info(f'sleeping {refresh_Rate}')
                    await asyncio.sleep(refresh_Rate)

            except Exception as error:
                logger.exception(error)

        @classmethod
        async def get(cls, group, faculty, session, delay):
            result = {}

            await asyncio.sleep(delay)
            
            logger.info(f'Refreshing group {group} on {datetime.now()}')

            try:
                # Находим ID факультета
                if Raspisanie.cache.get(faculty) is None:
                    faculty_Page = await Page.download(Raspisanie.groups_Link, session)
                    if faculty_Page == None:
                        return {'error': f'Не удалось скачать страницу с факультетами - {faculty}, {group}'}

                    faculty_ID = Page.find_Faculty_ID(faculty_Page, faculty)
                    if faculty_ID == None:
                        return {'error': f'Не удалось найти указанный факультет - {faculty}'}
                    else:
                        Raspisanie.cache[faculty] = faculty_ID
                else:
                    faculty_ID = Raspisanie.cache.get(faculty)

                # Находим ID группы
                if Raspisanie.cache.get(group) is None:
                    group_Page = await Page.download(f'{Raspisanie.groups_Link}{faculty_ID}', session)
                    if group_Page == None:
                        return {'error': f'Не удалось скачать страницу с расписанием групп - {faculty}, {group}'}

                    # Ищем нужную группу на странице
                    group_ID = Page.find_Group_ID(group_Page, group)
                    if group_ID == None:
                        return {'error': f'Не удалось найти указанную группу - {group}'}
                    else:
                        Raspisanie.cache[group] = group_ID
                else:
                    group_ID = Raspisanie.cache.get(group)

                # Получаем готовую ссылку на нужную группу
                link = f'{Raspisanie.groups_Link}{faculty_ID}/{group_ID}'

                try:
                    this_Week, next_Week = await Page.download_Rasp(link, session)
                except TypeError:
                    return {'error': f'Не нашел расписания для {group} группы'}
                
                # Парсим полученные странички
                this_Week = Page.parse(this_Week, False)
                next_Week = Page.parse(next_Week, False)

                # logger.debug(
                #     f'THIS WEEK GROUP={group}------------------------------ ' + '\n' + '\n' 
                #     + str(this_Week) + '\n' + '\n' +
                #     f'NEXT WEEK GROUP={group}------------------------------ ' + '\n' + '\n' 
                #     + str(next_Week) + '\n' + '\n'
                # )

                # Если парсер нашел нужное в страничках, то записываем их в переменные
                if this_Week != -10:
                    result.update(this_Week)
                if next_Week != -10:
                    result.update(next_Week)
                if this_Week == -10 and next_Week == -10:
                    return {'error': f'Не нашел расписания для {group} группы'}

                logger.info(f'Rasp for group {group} refreshed on {datetime.now()}')

                return {
                    'group' : group,
                    'payload': result
                }

            except Exception as error:
                logger.exception(error)

if __name__ == '__main__':
    Path = str(dirname(abspath(__file__))).rsplit('/', 1)[0]
    logger.add(f'{Path}/logs/log.log', backtrace = True, diagnose = True, format = '{time} {message}', level = 'DEBUG')

    with open(f'{Path}/configs/config.json', 'r') as config:
        cfg = loads(config.read())
        
    db = Database(Loop.loop, user = cfg['DB']['login'], passwd = cfg['DB']['password'])
    asyncio.run_coroutine_threadsafe(Raspisanie.init(Path, db, cfg['Refresher']['refresh_rate']), Loop.loop).result()

#!/usr/bin/python3

import asyncio
import aiohttp
from bs4 import BeautifulSoup
from datetime import datetime, timedelta
from loguru import logger
from random import SystemRandom
from re import compile
from time import sleep
from ssl import create_default_context

import errors
import pytypes
from db import Database

REFRESH_RATE = 259200

EXCLUDE_FACULTIES = (
    'ОБЩ',
    'АСП',
    'УРАИС',
    'ЦППК.',
    'МК',
    'АЛТГУ',
    'СПО',
    'ЭФ-В',
    'ФК',
    'ФПК'
)

EXCLUDE_GROUPS = (
    'шк',
    'школа фт'
    'цпс',
    'гос',
    'юристы'
)

class WebPage:
    """Представляет собой обычную web-страницу"""

    __headers = {
        ':authority': 'www.asu.ru',
        ':method': 'GET',
        ':scheme': 'https',
        'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9',
        'accept-Encoding': 'gzip, deflate, br',
        'accept-Language': 'ru,ru-RU;q=0.9,en-GB;q=0.8,en-US;q=0.7,en;q=0.6',
        'cache-Control': 'max-age=0',
        'dnt': '1',
        'sec-fetch-dest': 'document',
        'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36',
    }

    def __init__(self, session, link):
        self.session = session
        self.link = link
        
    async def get(self):
        """Скачивает и возвращает страницу"""

        WebPage.__headers[':path'] = f'/timetable/{self.link}'
        
        try:
            async with self.session.get(
                f'https://www.asu.ru/timetable/{self.link}', 
                headers=WebPage.__headers,
                ssl=create_default_context()
            ) as request:
                if request.status == 200:
                    return await request.text()
                else:
                    raise errors.PageDownloadError(f'http code - {request.status}')
        except Exception as error:
            raise errors.PageDownloadError(f'Can\'not download page, link - {self.link}, error - {error}')

class Cookie:
    """Представляет собой cookie файл"""

    def __init__(self, session):
        self.link = 'https://www.asu.ru/timetable/'
        self.session = session
    
    async def get(self) -> dict:
        """Получает куки и возвращает его"""

        async with self.session.get(
                self.link,
                ssl=create_default_context()
        ) as request:
            if request.status == 200:
                tmp = self.session.cookie_jar.filter_cookies(self.link)
                return {'__asu__': tmp['__asu__'].value}
            else:
                raise errors.PageDownloadError(f'http code - {request.status}')
        
class ContentPage(WebPage):
    """Представляет собой скаченную страницу с контентом"""

    def __init__(self, session, link):
        super().__init__(session, link)

    async def download(self, **kwargs):
        """Скачивает страницу и возвращает её.\n
        В случае status_code != 200 поднимает исключение BadPage/PageDownloadError \n
        kwargs: 
            delay(int) - sleep {delay} before download
            append - добавить к ссылке
        """

        data_to_append = kwargs.get('append', None)
        if data_to_append is not None:
            self.link += data_to_append

        delay = kwargs.get('delay', 0)
        if delay != 0:
            logger.debug(
                f'Sleeping {delay} seconds ' \
                f'before downloading {self.link}'
            )

            await asyncio.sleep(delay)

        logger.info(f'Downloading page - {self.link}')

        page = await self.get()
        if page is None:
            raise errors.BadPage

        return page

class FacultiesPage(ContentPage):
    """Представляет собой страницу со всеми факультетами вуза"""

    def __init__(self, session, link, **kwargs):
        """kwargs:
            delay (int) - задержка перед скачиванием страницы
        """

        super().__init__(session, link)

        self.delay = kwargs.get('delay', 0)

    async def find(self) -> list:
        """Скачивает страницу с факультетами и парсит её. \n
        Возвращает list со списком факультетов - pytypes.Faculty"""

        result = []
        pattern = compile(r'\((.*?)\)')

        soup = BeautifulSoup(
            await self.download(delay=self.delay),
            'html.parser'
        )
        content = soup.find_all(class_='link_ptr_left margin_bottom')
        
        for record in content:
            faculty_id = record.find('a', href=True)['href']
            faculty_name = pattern.search(record.text.rstrip().strip()).group(1)
            
            # Проверяем, нужно ли пропускать факультет
            if faculty_name not in EXCLUDE_FACULTIES \
                and faculty_name.rfind('-З') == -1:
                    result.append(
                        pytypes.Faculty(
                            id=faculty_id,
                            name=faculty_name,
                            link=f'{self.link}{faculty_id}'
                        )
                    )
            else:
                logger.debug(f'Skipping faculty {faculty_name}')
            
        return result

class GroupsPage(ContentPage):
    """Представляет собой страницу с ссылками на группы какого-либо факультета"""

    def __init__(self, session, faculty, **kwargs):
        """kwargs:
            delay (int) - задержка перед скачиванием страницы
        """

        super().__init__(session, f'students/{faculty.id}')

        self.faculty = faculty
        self.delay = kwargs.get('delay', 0)

    async def find(self) -> list:
        """Возвращает list с информацией о каждой группе факультета"""

        result = []
        soup = BeautifulSoup(
            await self.download(delay=self.delay), 
            'html.parser'
        )
        content = soup.find_all(class_='link_ptr_left margin_bottom')
        
        for record in content:
            site_id = record.find('a', href=True)['href']
            group_number = record.text.rstrip().strip()

            result.append(
                pytypes.Group(
                    id=site_id,
                    faculty=self.faculty,
                    num=group_number,
                    link=f'{self.link}{site_id}'
                )
            )
            
        return result

class GroupRaspPage(ContentPage):
    """Представляет собой страницу с расписанием какой-либо группы"""

    def __init__(self, session, group, **kwargs):
        super().__init__(session, group.link)

        self.group = group
        self.delay = kwargs.get('delay', 0)

    @property
    def group_info(self):
        """Возвращает информация о группе - pytypes.Group"""

        return self.group

    async def find(self):
        """Поиск расписания на странице"""

        result = {}

        # Вычисляем дату начала и конца некст недели
        start_next_week = datetime.now() + timedelta(days = 7 - datetime.now().weekday())
        end_next_week = start_next_week + timedelta(days = 6)

        try:
            # Скачиваем страницу с текущей и следующей неделей
            pages = (
                await self.download(delay=self.delay),
                await self.download(
                    delay=0.5,
                    append=f'?date={start_next_week.strftime("%Y%m%d")}-{end_next_week.strftime("%Y%m%d")}'
                )
            )
        except (errors.BadPage, errors.PageDownloadError):
            return -1

        for page in pages:
            try:
                soup = BeautifulSoup(page, 'html.parser')
                page = soup.find_all(class_='schedule-time')
            except IndexError:
                continue
            
            previous_time = ''

            for record in page:
                record = record.find_all('td')
                para_cleaned = ''
                data = None
            
                tmp = record[5].find('a', href=True)
                if tmp is None:
                    continue
                else:
                    tmp = str(tmp['href'])

                # У групп 56* и 57* в ссылке почему-то ;building
                # а у групп 58* и 59* в ссылке почему-то &building
                tmp = tmp.translate(str.maketrans({'&': '', ';': ''}))

                data = str(
                    datetime(
                        int(tmp[27:31]), 
                        int(tmp[31:33]), 
                        int(tmp[33:35])
                    ).date()
                )
            
                # Иногда на сайте не указана дата, просто пропускаем
                if data is None:
                    continue

                auditoriya = record[4].text.replace('\n', '').rstrip()
                para_name = record[2].text.replace('\n', '').split(' ')
                time = record[1].text.replace('\n', '').rstrip()

                # Если строка пуста (что бывает, когда у двух подгрупп пара в одно время)
                # то просто ставим предыдущее значение
                if time:
                    previous_time = time
                if not time:
                    time = previous_time

                for part in para_name:
                    if part != '':
                        para_cleaned += part + ' '

                if data not in result:
                    result[data] = {
                        'time': [time],
                        'para': [para_cleaned],
                        'auditoriya': [auditoriya]
                    }
                
                else:
                    result[data]['time'].append(time)
                    result[data]['para'].append(para_cleaned)
                    result[data]['auditoriya'].append(auditoriya)

        return self.group.faculty.name, self.group.num, result

class GroupsPagePoll:
    """Представляет собой множество страниц с ссылками на различные группы"""
    
    def __init__(self, session, faculties, rand_generator):
        self.session = session
        self.faculties = faculties
        self.generator = rand_generator

    async def get_all(self):
        """
            Скачивает все страницы факультетов 
            со списками групп и парсит скаченные страницы
        """

        tasks = []

        for faculty in self.faculties:
            delay = self.generator.uniform(0, len(self.faculties))

            tmp = GroupsPage(
                self.session,
                faculty,
                delay=delay
            )

            tasks.append(
                asyncio.create_task(
                    tmp.find()
                )
            )

        return await asyncio.gather(*tasks)
 
class GroupRaspPagePool:
    """Представляет собой множество страниц с расписанием различных групп"""
    
    def __init__(self, session, db_connection, pages, rand_generator):
        self.session = session
        self.pages = pages
        self.db = db_connection
        self.generator = rand_generator

    async def get_all(self):
        """
            Скачивает все страницы с расписанием и парсит их
        """

        tasks = []

        # Создаем экземпляры класса для каждой группы
        for group_list in self.pages:
            for group in group_list:

                # Проверяем, не нужно ли пропустить текущую группу
                if group.num.rfind('М') != -1 \
                    or group.num.rfind('асп') != -1 \
                    or group.num in EXCLUDE_GROUPS:
                        logger.debug(f'Skipping group {group.num}')
                        continue
                
                delay = self.generator.uniform(0, 5) + self.generator.uniform(0, 30)

                tmp = GroupRaspPage(
                    self.session,
                    group,
                    delay=delay
                )

                if len(tasks) < 10:
                    logger.debug(f'Creating task for group {group.num}, faculty - {group.faculty.name}')

                    tasks.append(
                        asyncio.create_task(
                            tmp.find()
                        )  
                    )
                else:
                    for group_rasp in await asyncio.gather(*tasks):
                        if group_rasp == -1:
                            logger.critical('Can\'not download page to parse')
                            continue
                        else:
                            faculty, group_num, payload = group_rasp

                        if not payload:
                            logger.critical(f'Payload doesn\'t contain any info - group = {group_num}, faculty - {faculty}')
                        else:
                            await self.db.insert(
                                faculty,
                                group_num, 
                                payload
                            )

                    tasks = []
                    delay = 32 - self.generator.uniform(0, 30)

                    await self.db.fsync()

                    logger.info(f'Sleeping {delay} seconds after downloading pages')
                    await asyncio.sleep(delay)

async def main():
    generator = SystemRandom()
    database = Database(asyncio.get_event_loop())
    
    async with aiohttp.TCPConnector(limit=0, ttl_dns_cache=300, keepalive_timeout=3000) as connector:
        logger.info('Getting cookie')

        async with aiohttp.ClientSession(connector=connector, connector_owner=False, cookie_jar=aiohttp.CookieJar()) as session:
            cookies = Cookie(session)
            site_cookie = await cookies.get()

        logger.debug(f'Cookies - {site_cookie}')
        await asyncio.sleep(generator.uniform(0, 5))
            
        async with aiohttp.ClientSession(connector=connector, connector_owner=False, cookies=site_cookie) as session:
            faculties_page = FacultiesPage(
                session,
                f'students/'
            )

            faculties = await faculties_page.find()
            
            groups_pages_pool = GroupsPagePoll(
                session,
                faculties,
                generator
            )

            groups_pages = await groups_pages_pool.get_all()

        delay = 60 + generator.uniform(0, 60) - generator.uniform(0, 60)
        logger.info(f'Sleeping after downloading pages - {delay} seconds')
        await asyncio.sleep(delay)

        async with aiohttp.ClientSession(connector=connector, connector_owner=False, cookies=site_cookie) as session:
            group_page_pool = GroupRaspPagePool(
                session,
                database,
                groups_pages,
                generator
            )

            await group_page_pool.get_all()
               
if __name__ == '__main__':
    logger.add(
        '/dev/null', backtrace=True, 
        diagnose=True, format='{time} {message}', 
        level='DEBUG'
    )

    while True:
        try:
            loop = asyncio.get_event_loop()
            loop.run_until_complete(main())
        finally:
            loop.close()

        logger.info(f'Finished parsing on {datetime.now()}, sleeping {REFRESH_RATE}')
        sleep(REFRESH_RATE)

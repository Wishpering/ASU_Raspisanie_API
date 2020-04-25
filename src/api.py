#!/usr/bin/python3

from os.path import dirname, abspath
from socket import gethostname, gethostbyname
from fastapi import FastAPI, Header, HTTPException
from uvicorn import run
from loguru import logger
from secrets import token_hex
from json import loads
from datetime import datetime
import asyncio

from db import Database
from date import Date_Utils

class Api:
    app = FastAPI(title = 'Asu raspisanie API', description = 'API для получения расписания с сайта asu.ru')
    link_To_DB = None
    password = None
    
    @app.on_event("startup")
    async def on_Startup():
        Path = str(dirname(abspath(__file__))).rsplit('/', 1)[0]
        
        logger.add(f'{Path}/logs/log.log', colorize = True, backtrace = True, diagnose = True, format = '{time} {message}', level = 'DEBUG')

        with open(f'{Path}/configs/config.json', 'r') as config:
            cfg = loads(config.read())
        
        Api.password = cfg['Api']['password']
        Api.link_To_DB = Database(asyncio.get_running_loop(), user = cfg['DB']['login'], passwd = cfg['DB']['password'])
        
    @app.get("/")
    async def root(authorization: str = Header(None)):
        # Проверяем токен из headers
        if authorization is None or await Api.link_To_DB.check_Token(authorization) != True:
            raise HTTPException(status_code = 401, detail = 'you are not welcome here')
        else: 
            return {'payload' : 'Nothing here'}

    @app.get("/rasp")
    async def get_Rasp(prep : str = None, group : str = None, date : str = None, end_date : str = None, authorization: str = Header(None)):
        # Проверяем токен из headers
        if authorization is None or await Api.link_To_DB.check_Token(authorization) != True:
            raise HTTPException(status_code = 401, detail = 'you are not welcome here')
        else:
            if prep is not None and group is not None:
                raise HTTPException(400)

            if prep is not None:
                # Расписание всех преподавателей в пуле
                if prep == 'all':
                    return await Api.link_To_DB.listall_Preps()

                # Имена преподавателей в пуле и их количество
                elif prep == 'pool':
                    return await Api.link_To_DB.preps_Pool()
                
                # Расписание какого-то конкретного преподавателя
                elif await Api.link_To_DB.search_Prep(prep) is not None:
                    search_Res = await Api.link_To_DB.search_Prep(prep)

                    # Расписание на конкретный день
                    if date is not None and end_date is None:
                        if search_Res.get(date) is not None:
                            return search_Res.get(date)
                        else:
                            raise HTTPException(400)

                    # Расписание на период
                    elif end_date is not None:
                        result = {}

                        if date is None:
                            date, end_date = datetime.now(), Date_Utils.to_Datetime(end_date)
                        else:
                            date, end_date = Date_Utils.to_Datetime(date, end_date)

                        if date is None or end_date is None:
                            raise HTTPException(400)

                        for date in Date_Utils.data_Range(date, end_date):
                            tmp = search_Res.get(f'{date.date()}')

                            if tmp is not None:
                                result[date.date()] = tmp
                            else:
                                result[date.date()] = None

                        return result

                    # Все расписание преподавателя
                    elif date is None and end_date is None:
                        return search_Res

                    else:
                        raise HTTPException(404)

                else:
                    raise HTTPException(404)

            if group is not None:
                # Список групп в пуле и их расписание
                if group == 'all':
                    return await Api.link_To_DB.listall()

                # Список групп в пуле и их количество
                elif group == 'pool':
                    return await Api.link_To_DB.groups_Pool()

                if await Api.link_To_DB.search(group) is not None:
                    search_Res = await Api.link_To_DB.search(group)

                    # Расписание на конкретный день
                    if date is not None and end_date is None:
                        if search_Res.get(date) is not None:
                            return search_Res.get(date)
                        else:
                            raise HTTPException(404)

                    # Расписание на период
                    elif end_date is not None:
                        result = {}

                        if date is None:
                            date, end_date = datetime.now(), Date_Utils.to_Datetime(end_date)
                        else:
                            date, end_date = Date_Utils.to_Datetime(date, end_date)

                        if date is None or end_date is None:
                            raise HTTPException(400)
                
                        for date in Date_Utils.data_Range(date, end_date):
                            tmp = search_Res.get(f'{date.date()}')

                            if tmp is not None:
                                result[date.date()] = tmp
                            else:
                                result[date.date()] = None

                        return result

                    # Все расписание группы
                    elif date is None and end_date is None:
                        return search_Res

                    else:
                        raise HTTPException(404, 'Not found')

                else:
                    raise HTTPException(404, 'Not found')

            else:   
                raise HTTPException(404, 'Not found')

    @app.get('/token')
    async def generate_Token(password : str = None):
        # Если пароль неверный, то не даем создать токен
        if password is None or password != Api.password:
            raise HTTPException(403, 'Wrong password')
        else:
            # Генерируем токен и пытаемся записать его в бд
            token = token_hex(10)
            db_Response = await Api.link_To_DB.insert_Token(token)

            # Если каким-то чудом токен уже совпадает с каким-то,
            # то генерим новый и проверяем заново
            while db_Response == -1:
                token = token_hex(10)
                db_Response = await Api.link_To_DB.insert_Token(token)

            return {'token' : token}

if __name__ == '__main__':
    run("api:Api.app", host = gethostbyname(gethostname()) , port = 80, log_level = "debug", reload = True)

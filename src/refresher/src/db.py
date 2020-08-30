import motor.motor_asyncio

class Database:
    options = {'async': True}

    def __init__(self, loop, **kwargs):
        """Создать соединение с БД"""
        """
            loop = asyncio.loop
            kwargs:
                address - адрес БД, default = database
                port - порт БД, default = 27017
                user - юзер для бд, default = None
                passwd - пароль для юзера, default = None
        """

        db_Address = kwargs.get('address', 'database')
        db_Port = kwargs.get('port', 27017)
        db_User = kwargs.get('user', None)
        db_Passwd = kwargs.get('passwd', None)

        self.db_connection = motor.motor_asyncio.AsyncIOMotorClient(
            f'mongodb://{db_Address}:{db_Port}/', 
            username = db_User, 
            password = db_Passwd,
            io_loop = loop
        )

        self.db = self.db_connection['rasp']

    async def fsync(self):
        """Вызвать fsync"""

        await self.db_connection.fsync(**Database.options)

    async def insert(self, faculty, group, payload):
        """Вставить расписание группы в БД"""
        """Параметры - group = str, faculty = str, payload = json"""

        self.collection = self.db[faculty]

        if await self.collection.count_documents({'group': group}) >= 1:
            await self.collection.replace_one(
                {'group' : group},
                {
                    'group' : group, 
                    'payload' : payload
                }
            )

        else:
            await self.collection.insert_one(
                {
                    'group' : group, 
                    'payload' : payload
                }
            )
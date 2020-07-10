import motor.motor_asyncio

class Database:
    options = {'async': True}

    def __init__(self, loop, **kwargs):
        db_Address = kwargs.get('address', 'database')
        db_Port = kwargs.get('port', 27017)
        db_User = kwargs.get('user', None)
        db_Passwd = kwargs.get('passwd', None)

        self.db_Client = motor.motor_asyncio.AsyncIOMotorClient(
            f'mongodb://{db_Address}:{db_Port}/', 
            username = db_User, 
            password = db_Passwd,
            io_loop = loop
        )

        self.groups = self.db_Client['Raspisanie']['groups']
        self.preps = self.db_Client['Raspisanie']['preps'] 
        self.tokens = self.db_Client['Tokens']['tokens']

    async def fsync(self):
        await self.db_Client.fsync(**Database.options)

    async def insert(self, group, payload):
        if await self.groups.count_documents({"group": group}) >= 1:
            await self.groups.replace_one(
                {'group' : group},
                {
                    'group' : group, 
                    'payload' : payload
                }
            )

        else:
            await self.groups.insert_one(
                {
                    'group' : group, 
                    'payload' : payload
                }
            )

    async def insert_Prep(self, prep, payload):
        if await self.preps.count_documents({"prep": prep}) >= 1:
            await self.preps.replace_one(
                {'prep' : prep},
                {
                    'prep' : prep, 
                    'payload' : payload
                }
            )

        else:
            await self.preps.insert_one(
                {
                    'prep' : prep, 
                    'payload' : payload
                }
            )

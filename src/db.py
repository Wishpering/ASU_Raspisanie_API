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

    async def listall_Preps(self):
        result = {}

        async for record in self.preps.find():
            result[record.get('prep')] = record.get('payload')

        return result

    async def listall(self):
        result = {}

        async for record in self.groups.find():
            result[record.get('group')] = record.get('payload')

        return result

    async def search_Prep(self, prep):
        async for record in self.preps.find({'prep' : {'$regex' : prep.lower()}}):
            return record.get('payload')

    async def search(self, group):
        async for record in self.groups.find({'group' : group}):  
            return record.get('payload')

    async def preps_Pool(self):
        result = []

        async for record in self.preps.find():
            tmp = record.get('prep')

            if tmp not in result:
                result.append(tmp)

        return {
            'count' : len(result),
            'payload' : result
        }

    async def groups_Pool(self):
        result = []

        async for record in self.groups.find():
            tmp = record.get('group')

            if tmp not in result:
                result.append(tmp)

        return {
            'count' : len(result),
            'payload' : result
        }

    async def insert_Token(self, token):
        if await self.tokens.count_documents({'token' : token}) >= 1:
            return -1

        else:
            await self.tokens.insert_one(
                {
                   'token' : token
                }
            )
            
            await self.db_Client.fsync(**Database.options)

    async def check_Token(self, token):
        if await self.tokens.count_documents({'token' : token}) == 1:
            return True
        else:
            return False
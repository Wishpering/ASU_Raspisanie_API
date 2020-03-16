import pymongo

class Database:
    def __init__(self, **kwargs):
        db_Address = kwargs.get('address', 'database')
        db_Port = kwargs.get('port', 27017)
        db_User = kwargs.get('user', None)
        db_Passwd = kwargs.get('passwd', None)

        db_Client = pymongo.MongoClient(f'mongodb://{db_Address}:{db_Port}/', username = db_User, password = db_Passwd)
        self.groups = db_Client['Raspisanie']['groups']
        self.preps = db_Client['Raspisanie']['preps'] 
        self.tokens = db_Client['Tokens']['tokens']

    def insert(self, group, payload):
        if self.groups.count_documents({"group": group}) >= 1:
            self.groups.replace_one(
                {'group' : group},
                {
                    'group' : group, 
                    'payload' : payload
                }
            )

        else:
            self.groups.insert_one(
                {
                    'group' : group, 
                    'payload' : payload
                }
            )

    def insert_Prep(self, prep, payload):
        if self.preps.count_documents({"prep": prep}) >= 1:
            self.preps.replace_one(
                {'prep' : prep},
                {
                    'prep' : prep, 
                    'payload' : payload
                }
            )

        else:
            self.preps.insert_one(
                {
                    'prep' : prep, 
                    'payload' : payload
                }
            )

    def listall_Preps(self):
        result = {}

        for record in self.preps.find():
            result[record.get('prep')] = record.get('payload')

        return result

    def listall(self):
        result = {}

        for record in self.groups.find():
            result[record.get('group')] = record.get('payload')

        return result

    def search_Prep(self, prep):
        for record in self.preps.find({'prep' : {'$regex' : prep.lower()}}):
            return record.get('payload')

    def search(self, group):
        for record in self.groups.find({'group' : group}):  
            return record.get('payload')

    def preps_Pool(self):
        result = []

        for record in self.preps.find():
            tmp = record.get('prep')

            if tmp not in result:
                result.append(tmp)

        return {
            'count' : len(result),
            'payload' : result
        }

    def groups_Pool(self):
        result = []

        for record in self.groups.find():
            tmp = record.get('group')

            if tmp not in result:
                result.append(tmp)

        return {
            'count' : len(result),
            'payload' : result
        }

    def insert_Token(self, token):
        if self.tokens.count_documents({'token' : token}) >= 1:
            return -1

        else:
            self.tokens.insert_one(
                {
                   'token' : token
                }
            )

    def check_Token(self, token):
        if self.tokens.count_documents({'token' : token}) >= 1:
            return True
        else:
            return False

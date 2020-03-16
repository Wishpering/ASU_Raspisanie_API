from datetime import datetime, timedelta

class Date_Utils:
    __translate_Table = str.maketrans({'-' : '', ':' : ''})

    @classmethod
    def to_Datetime(cls, *args):
        result = []

        for date in args:            
            date = date.translate(Date_Utils.__translate_Table)

            if len(date) == 8:
                try:
                    date = datetime(int(date[0:4]), int(date[4:6]), int(date[6:8]))
                except ValueError:
                    date = datetime(int(date[4:8]), int(date[2:4]), int(date[0:2]))

                result.append(date)
            else:
                result.append(None)

        if len(result) == 0:
            return None
        elif len(result) == 1:
            return result[0]
        else:
            return result

    @classmethod
    def data_Range(cls, start, end):  
        # Иначе будем получать неверный результат
        start = start.replace(hour = 0, minute = 0, second = 0, microsecond = 0)
        end = end.replace(hour = 0, minute = 0, second = 0, microsecond = 0)

        for i in range(0, (end - start).days + 1):
            yield start + timedelta(days = i)

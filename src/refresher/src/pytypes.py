from pydantic import BaseModel

class Faculty(BaseModel):
    """
    id - id факультета на сайте \n
    name - название факультета \n
    link - ссылка на факультет
    """

    id: str
    name: str
    link: str

class Group(BaseModel):
    """
    id - id группы на сайте \n
    faculty - pytypes.Faculty \n
    num - название группы (575.../etc) \n
    link - ссылка на группу
    """

    id: str
    faculty: Faculty
    num: str
    link: str
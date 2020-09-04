from dataclasses import dataclass

@dataclass
class Faculty:
    """
    id - id факультета на сайте \n
    name - название факультета \n
    link - ссылка на факультет
    """

    id: str
    name: str
    link: str

@dataclass
class Group:
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

@dataclass
class Raspisanie:
    """
    faculty - имя факультета \n
    group - название группы (575.../etc) \n
    rasp - расписание
    """

    faculty: str
    group: str
    rasp: dict
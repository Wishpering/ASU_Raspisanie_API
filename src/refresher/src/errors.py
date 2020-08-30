class BadPage(Exception):
    pass

class PageDownloadError(Exception):
    def __init__(self, error_info):
        self.info = error_info

    def __str__(self):
        return(repr(self.info))

class ApiRef:
    def __init__(self, api: str, name: str):
        self.api = api
        self.name = name
        self.combined = format("{}:{}".format(api, name))
    def __eq__(self, other):
        return self.combined.__eq__(other.combined)
    def __ge__(self, other):
        return self.combined.__ge__(other.combined)
    def __lt__(self, other):
        return self.combined.__lt__(other.combined)
    def __hash__(self):
        return self.combined.__hash__()
    def __str__(self):
        return self.combined
    def __repr__(self):
        return self.combined

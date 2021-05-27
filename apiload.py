import os
import sys

BRANCH = "10.0.19041.202-preview"
SHA = "7164f4ce9fe17b7c5da3473eed26886753ce1173"

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
WIN32JSON_REPO = os.path.join(SCRIPT_DIR, "win32json")
WIN32JSON_API_DIR = os.path.join(WIN32JSON_REPO, "api")

def loadSortedList():
    if not os.path.exists(WIN32JSON_REPO):
        print("Error: missing '{}'".format(WIN32JSON_REPO))
        print("Clone it with:")
        print("  git clone https://github.com/marlersoft/win32json {} -b {} && git -C {} checkout {} -b for_win32_transform".format(WIN32JSON_REPO, win32json_branch, win32json, win32json_sha))
        sys.exit(1)

    def getApiName(basename: str) -> str:
        if not basename.endswith(".json"):
            sys.exit("found a non-json file '{}' in directory '{}'".format(basename, WIN32JSON_API_DIR))
        return basename[:-5]
    apis = [getApiName(basename) for basename in os.listdir(WIN32JSON_API_DIR)]

    # just always sort to aid in predictable output
    apis.sort()

    return apis

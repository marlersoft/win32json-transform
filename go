#!/usr/bin/env python3
#
# Steps
# --------------------------------------------------------------------------------
# * identify all the "direct ApiRef dependencies" for every type
# * using these direct dependnecies, identify which types have dependencies
#   that escape the current Api, then come back
# * all types in this *cycle*, should be moved to a core json file

import os
import sys
import json
from typing import List, Set, Dict

def apiRefTypeName(type_obj):
    return ".".join(type_obj["Parents"] + [type_obj["Name"]])

#def apiRefName(type_obj):
#    return type_obj["Api"] + ":" + apiRefTypeName(type_obj)

class DefaultDict(dict):
    def __init__(self, factory):
        self.factory = factory
    def __missing__(self, key):
        self[key] = self.factory(key)
        return self[key]


#class TypeRef:
#    def __init__(self, referenced_by, type_reference):
#        self.referenced_by = referenced_by
#        self.type_reference = type_reference
class TypeRefSet:
    def __init__(self, name: str):
        self.name = name
        self.referenced_by: List[str] = []

class ApiRefSet:
    def __init__(self, name: str):
        self.name = name
        #self.refs: List[TypeRef] = []
        self.refs = DefaultDict(TypeRefSet)

class ApiObj:
    def __init__(self, json_obj: dict):
        self.json_obj = json_obj
class ApiObjConstant(ApiObj):
    def __init__(self, json_obj: dict):
        super().__init__(json_obj)


class ApiRef:
    def __init__(self, api: str, name: str):
        self.api = api
        self.name = name
        self.combined = format("{}:{}".format(api, name))
    def __eq__(self, other):
        return self.combined.__eq__(other.combined)
    def __hash__(self):
        return self.combined.__hash__()

def makeApiSet(ignore) -> Set[str]:
    return set()

def getJsonApiRefs(api_refs: Set[ApiRef], json_obj):
    if isinstance(json_obj, dict):
        if "Kind" in json_obj:
            if json_obj["Kind"] == "ApiRef":
                api_refs.add(ApiRef(json_obj["Api"], apiRefTypeName(json_obj)))
                return
        for name,val in json_obj.items():
            getJsonApiRefs(api_refs, val)
    elif isinstance(json_obj, list):
        for entry in json_obj:
            getJsonApiRefs(api_refs, entry)
    elif isinstance(json_obj, str) or (json_obj is None) or isinstance(json_obj, bool) or isinstance(json_obj, int):
        pass
    else:
        sys.exit("{}".format(type(json_obj).__name__))


def isAnonType(type_name):
    return type_name.endswith("_e__Struct") or type_name.endswith("_e__Union")

def main():
    win32json_branch = "10.0.19041.202-preview"
    win32json_sha = "7164f4ce9fe17b7c5da3473eed26886753ce1173"

    script_dir = os.path.dirname(os.path.abspath(__file__))
    win32json = os.path.join(script_dir, "win32json")
    if not os.path.exists(win32json):
        print("Error: missing '{}'".format(win32json))
        print("Clone it with:")
        print("  git clone https://github.com/marlersoft/win32json {} -b {} && git -C {} checkout {} -b for_win32_transform".format(win32json, win32json_branch, win32json, win32json_sha))

    api_dir = os.path.join(win32json, "api")

    def getApiName(basename):
        if not basename.endswith(".json"):
            sys.exit("found a non-json file '{}' in directory '{}'".format(basename, api_dir))
        return basename[:-5]
    apis = [getApiName(basename) for basename in os.listdir(api_dir)]
    apis.sort()

    #api_deps = ApiDeps()
    api_direct_deps = DefaultDict(makeApiSet)

    api_direct_type_refs_table: dict[str,dict[str,Set[ApiRef]]] = {}
    print("loading types...")
    for api_name in apis:
        #print(api_name)
        filename = os.path.join(api_dir, api_name + ".json")
        with open(filename, "r", encoding='utf-8-sig') as file:
            api = json.load(file)
        constants = api["Constants"]
        types = api["Types"]
        functions = api["Functions"]
        unicode_aliases = api["UnicodeAliases"]

        api_ref_sets = DefaultDict(ApiRefSet)

        #print("  {} Constants".format(len(constants)))

#        types = {}
#        for constant in constants:
#            constant_type = constant["Type"]
#            kind = constant_type["Kind"]
#            if kind == "ApiRef":
#                api_entry = api_ref_sets[constant_type["Api"]]
#                ref_name = apiRefTypeName(constant_type)
#                api_entry.refs[ref_name].referenced_by.append(ApiObjConstant(constant))
#        for api_ref_set in api_ref_sets.values():
#            print("    Referencing {} types from '{}'".format(len(api_ref_set.refs), api_ref_set.name))
#            api_direct_deps[api_name].add(api_ref_set.name)
#            for ref_name in api_ref_set.refs:
#                print("      {} referenced by {} types".format(ref_name, len(api_ref_set.refs[ref_name].referenced_by)))


        #print("  {} Types".format(len(types)))
        #print("  {} Functions".format(len(functions)))
        #print("  {} UnicodeAliases".format(len(unicode_aliases)))

        def addTypeRefs(type_refs_table: Dict[str,Set[ApiRef]], name_prefix: str, type_json: Dict):
            kind = type_json["Kind"]
            full_name = name_prefix + type_json["Name"]
            #if len(name_prefix) > 0:
            #    print("nested type '{}'".format(full_name))
            api_refs: Set[ApiRef] = set()
            if kind == "Struct" or kind == "Union":
                getJsonApiRefs(api_refs, type_json)
                for nested_type_json in type_json["NestedTypes"]:
                    addTypeRefs(type_refs_table, full_name + ".", nested_type_json)
            elif kind == "Enum" or kind == "ComClassID":
                getJsonApiRefs(api_refs, type_json)
                assert(len(api_refs) == 0)
            elif kind == "Com" or kind == "FunctionPointer" or kind == "NativeTypedef":
                getJsonApiRefs(api_refs, type_json)
            else:
                sys.exit(kind)

            # this could happen because of the same types defined for multiple architectures
            if full_name in type_refs_table:
                #raise ValueError("key '{}' already exists".format(full_name))
                type_refs_table[full_name] = type_refs_table[full_name].union(api_refs)
            else:
                type_refs_table[full_name] = api_refs


        type_refs_table: dict[str,Set[ApiRef]] = {}
        for type_json in types:
            addTypeRefs(type_refs_table, "", type_json)
        api_direct_type_refs_table[api_name] = type_refs_table

    print("types loaded")
    for api in apis:
        direct_type_refs_table = api_direct_type_refs_table[api]
        print("api {}".format(api))

        i = 0
        for type_name,refs in direct_type_refs_table.items():
            i += 1
            if refs:
                print("    {} refs from {}".format(len(refs), type_name))
                for ref in refs:
                    #print("        {}".format(ref.combined))
                    # make sure the API is valid
                    table = api_direct_type_refs_table[ref.api]
                    # if not an anonymous type, make sure the type is valid
                    if not isAnonType(ref.name):
                        if ref.name not in table:
                            # check if it is a nested type here
                            # TODO: I think I need to check every level of nesting
                            print("api {} type_name{} ref.name {}".format(api, type_name, ref.name))
                            type_names = type_name.split(".")
                            ref_names = ref.name.split(".")
                            i = 0
                            while i < len(ref_names) and i < len(type_names):
                                if ref_names[i] != type_names[-i-1]:
                                    break
                                i += 1

                            nested_name = ".".join(type_names + ref_names[i:])
                            if not nested_name in direct_type_refs_table:
                                sys.exit("!!!!!!!!!!! {} not in {} (and {} not in {})".format(ref.name, api, nested_name, api))




#    recursive_dep_table: dict[str,Set[str]] = {}
#    for api in apis:
#        recursive_deps: Set[str] = set()
#        getRecursiveDeps(api_direct_deps, api, recursive_deps)
#        #print("{}: {}".format(api, recursive_deps))
#        recursive_dep_table[api] = recursive_deps
#    for api in apis:
#        recursive_deps = recursive_dep_table[api]
#        print("{}: {}".format(api, recursive_deps))
#        if api in recursive_deps:
#            print("{} is CYCLIC!!!!".format(api))
#
#def getRecursiveDeps(direct_dep_table: Set[Set[str]], name: str, result: Set[str]) -> None:
#    for direct_dep in direct_dep_table[name]:
#        if not direct_dep in result:
#            result.add(direct_dep)
#            getRecursiveDeps(direct_dep_table, direct_dep, result)
#

main()
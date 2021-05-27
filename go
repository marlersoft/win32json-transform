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
from typing import List, Set, Dict, Optional

def getApiRefTopLevelType(type_obj):
    parents = type_obj["Parents"]
    if parents:
        return parents[0]
    return type_obj["Name"]

class DefaultDict(dict):
    def __init__(self, factory):
        self.factory = factory
    def __missing__(self, key):
        self[key] = self.factory(key)
        return self[key]

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
class ApiTypeNameToApiRefMap:
    def __init__(self):
        self.top_level: Dict[str,Set[ApiRef]] = {}
        self.nested: Dict[str,Set[ApiRef]] = {}

def getJsonApiRefs(api_refs: Set[ApiRef], json_obj):
    if isinstance(json_obj, dict):
        if "Kind" in json_obj:
            if json_obj["Kind"] == "ApiRef":
                api_refs.add(ApiRef(json_obj["Api"], getApiRefTopLevelType(json_obj)))
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

    def getApiName(basename: str) -> str:
        if not basename.endswith(".json"):
            sys.exit("found a non-json file '{}' in directory '{}'".format(basename, api_dir))
        return basename[:-5]
    apis = [getApiName(basename) for basename in os.listdir(api_dir)]
    apis.sort()

    api_direct_type_refs_table: Dict[str,ApiTypeNameToApiRefMap] = {}
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

        #api_ref_sets = DefaultDict(ApiRefSet)

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

        def addTypeRefs(type_map: ApiTypeNameToApiRefMap, name_prefix: str, type_json: Dict):
            kind = type_json["Kind"]
            full_name = name_prefix + type_json["Name"]
            #if len(name_prefix) > 0:
            #    print("nested type '{}'".format(full_name))
            api_refs: Set[ApiRef] = set()
            if kind == "Struct" or kind == "Union":
                getJsonApiRefs(api_refs, type_json["Fields"])
                for nested_type_json in type_json["NestedTypes"]:
                    addTypeRefs(type_map, full_name + ".", nested_type_json)
            elif kind == "Enum" or kind == "ComClassID":
                getJsonApiRefs(api_refs, type_json)
                assert(len(api_refs) == 0)
            elif kind == "Com" or kind == "FunctionPointer" or kind == "NativeTypedef":
                getJsonApiRefs(api_refs, type_json)
            else:
                sys.exit(kind)

            # this could happen because of the same types defined for multiple architectures
            table = type_map.top_level if (len(name_prefix) == 0) else type_map.nested
            if full_name in table:
                #raise ValueError("key '{}' already exists".format(full_name))
                table[full_name] = table[full_name].union(api_refs)
            else:
                table[full_name] = api_refs


        type_map = ApiTypeNameToApiRefMap()
        for type_json in types:
            addTypeRefs(type_map, "", type_json)
        api_direct_type_refs_table[api_name] = type_map

    print("types loaded, checking that type refs exist...")
    for api in apis:
        direct_type_refs_table = api_direct_type_refs_table[api]
        #print("api {}".format(api))

        i = 0
        for type_name,refs in direct_type_refs_table.top_level.items():
            i += 1
            if refs:
                #print("    {} refs from {}".format(len(refs), type_name))
                for ref in refs:
                    #print("        {}".format(ref.combined))
                    # make sure the API is valid
                    table = api_direct_type_refs_table[ref.api]
                    # if not an anonymous type, make sure the type is valid
                    if not isAnonType(ref.name):
                        if ref.name not in table.top_level:
                            nested_name = getNestedName(type_name, ref)
                            if not nested_name in direct_type_refs_table.nested:
                                sys.exit("!!!!!!!!!!! {} not in {} (and {} not in {})".format(ref.name, api, nested_name, api))
                            # TODO: should we save this nested name? not sure that we need to

    print("types verified")

    print("calculating recursive type references...")
    api_recursive_type_refs_table: Dict[str,dict[str,Set[ApiRef]]] = {}
    with open(os.path.join(script_dir, "out-recursive-deps.txt"), "w") as file:
        for api in apis:
            #print("calculating recursive deps on {}...".format(api))
            direct_type_refs_table = api_direct_type_refs_table[api]
            recursive_type_refs_table = {}
            for type_name,refs in direct_type_refs_table.top_level.items():
                recursive_chains: List[List[ApiRef]] = []
                getRecursiveChains(api_direct_type_refs_table, set(), refs, recursive_chains, None)
                file.write("{}:{} -> {}\n".format(api, type_name, recursive_chains))
                recursive_type_refs_table[type_name] = recursive_chains
            api_recursive_type_refs_table[api] = recursive_type_refs_table
    print("done calculating recursive type references")

    print("searching for cycles...")
    full_type_set: Set[ApiRef] = set()
    cycle_type_set: Set[ApiRef] = set()
    with open(os.path.join(script_dir, "out-cycles.txt"), "w") as file:
        for api in apis:
            #print("API: {}".format(api))
            table = api_recursive_type_refs_table[api]
            cycle_count = 0
            for type_name, recursive_chains in table.items():
                type_api_ref = ApiRef(api, type_name)
                full_type_set.add(type_api_ref)
                for chain in recursive_chains:
                    state = 0
                    for ref in chain:
                        if ref.api == api:
                            if state == 1:
                                state = 2
                                break
                        else:
                            if state == 0:
                                state = 1
                    if state == 2:
                        file.write("{}:{}  CHAIN={}\n".format(api, type_name, chain))
                        cycle_count += 1
                        cycle_type_set.add(type_api_ref)
                        for ref in chain:
                            cycle_type_set.add(ref)
                    else:
                        pass
                        #print("NOT CYCLIC: {}:{}  CHAIN={}".format(api, type_name, chain))
            if cycle_count > 0:
                print("{:4} cycles: {}:{}".format(cycle_count, api, type_name))

    print("{} out of {} types involved in cycles".format(len(cycle_type_set), len(full_type_set)))
    with open(os.path.join(script_dir, "out-cycle-types.txt"), "w") as file:
        cycle_types_list = list(cycle_type_set)
        cycle_types_list.sort()
        for cycle_type in cycle_types_list:
            file.write("{}\n".format(cycle_type.combined))

    print("done")


def getRecursiveChains(api_direct_type_refs_table: Dict[str,ApiTypeNameToApiRefMap], handled: Set[ApiRef], refs: Set[ApiRef], result: List[List[ApiRef]], current_chain: Optional[List[ApiRef]]) -> None:
    for ref in refs:
        ref_api_table = api_direct_type_refs_table[ref.api]
        if not isAnonType(ref.name) and ref.name in ref_api_table.top_level:
            #file.write("\"{}\" -> \"{}\";\n".format(type_name, ref.name))
            next_chain = current_chain
            if not next_chain:
                next_chain = []
                result.append(next_chain)
            next_chain.append(ref)

            ref_refs = ref_api_table.top_level[ref.name]
            if not ref in handled:
                handled.add(ref)
                getRecursiveChains(api_direct_type_refs_table, handled, ref_refs, result, next_chain)

def getNestedName(type_name: str, ref: ApiRef) -> str:
    type_names = type_name.split(".")
    ref_names = ref.name.split(".")
    i = 0
    while i < len(ref_names) and i < len(type_names):
        if ref_names[i] != type_names[-i-1]:
            break
        i += 1
    return  ".".join(type_names + ref_names[i:])


main()

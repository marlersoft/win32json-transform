
def writeInline(writer, json_val):
    if json_val is None:
        writer("null")
    elif isinstance(json_val, dict):
        writer("{")
        writeInlineFields(writer, json_val.keys(), "", "", json_val)
        writer("}")
    elif isinstance(json_val, list):
        writer("[")
        prefix = ""
        for val in json_val:
            writer(prefix)
            writeInline(writer, val)
            prefix = ","
        writer("]")
    elif isinstance(json_val, bool):
        writer("true" if json_val else "false")
    elif isinstance(json_val, str):
        writer("\"{}\"".format(json_val))
    elif isinstance(json_val, int):
        writer("{}".format(json_val))
    #if (json_val is None) or isinstance(json_val, bool) or isinstance(json_val, int):
    #    pass
    else:
        import sys
        sys.exit("{}".format(type(json_val).__name__))


def writeFieldName(writer, field_prefix, opt_comma, name):
        writer("{}{}\"{}\":".format(field_prefix, opt_comma, name))

def writeInlineFields(writer, fields, field_prefix, field_suffix, obj, separator=""):
    for field in fields:
        writeFieldName(writer, field_prefix, separator, field)
        writeInline(writer, obj[field])
        writer(field_suffix)
        separator = ","

def writeInlineElements(writer, line_prefix, arr, separator=""):
    for element in arr:
        writer("\r\n{}{}".format(line_prefix, separator))
        writeInline(writer, element)
        separator = ","

def writeConstant(writer, prefix, constant):
    writer("\t{}{{\r\n".format(prefix))
    writeInlineFields(writer, ["Name","Type","ValueType","Value","Attrs"], "\t\t", "\r\n", constant)
    writer("\t}\r\n")

def writeComMethod(writer, method_separator, method):
    line_prefix = "\t\t\t"

    writer("{}{}{{\r\n".format(line_prefix, method_separator))
    child_prefix = line_prefix + "\t"
    writeInlineFields(writer, ["Name", "SetLastError", "ReturnType", "Architectures", "Platform", "Attrs"], child_prefix, "\r\n", method)

    writeFieldName(writer, child_prefix, ",", "Params")
    writer("[")
    writeInlineElements(writer, "\t\t\t\t\t", method["Params"])
    writer("\r\n{}]\r\n".format(child_prefix))

    writer("{}}}\r\n".format(line_prefix))

def writeType(writer, line_prefix, type_separator, type_obj):
    writer("{}{}{{\r\n".format(line_prefix, type_separator))
    writeInlineFields(writer, ["Name", "Architectures", "Platform", "Kind"], line_prefix + "\t", "\r\n", type_obj)
    kind = type_obj["Kind"]
    child_prefix = line_prefix + "\t"
    if kind == "Enum":
        writeInlineFields(writer, ["Flags", "Scoped"], child_prefix, "\r\n", type_obj, separator=",")

        writeFieldName(writer, child_prefix, ",", "Values")
        writer("[")
        writeInlineElements(writer, child_prefix + "\t", type_obj["Values"])
        writer("\r\n{}]\r\n".format(child_prefix))

        writeInlineFields(writer, ["IntegerBase"], child_prefix, "\r\n", type_obj, separator=",")
    elif kind == "Struct" or kind == "Union":
        writeInlineFields(writer, ["Size", "PackingSize"], child_prefix, "\r\n", type_obj, separator=",")

        writeFieldName(writer, child_prefix, ",", "Fields")
        writer("[")
        writeInlineElements(writer, child_prefix + "\t", type_obj["Fields"])
        writer("\r\n{}]\r\n".format(child_prefix))

        writeFieldName(writer, child_prefix, ",", "NestedTypes")
        writer("[\r\n")
        nested_separator = ""
        for nested_type in type_obj["NestedTypes"]:
            writeType(writer, child_prefix, nested_separator, nested_type)
            nested_separator=","
        writer("{}]\r\n".format(child_prefix))

    elif kind == "Com":
        writeInlineFields(writer, ["Guid", "Interface"], child_prefix, "\r\n", type_obj, separator=",")

        writeFieldName(writer, child_prefix, ",", "Methods")
        writer("[\r\n")
        method_separator = ""
        for method in type_obj["Methods"]:
            writeComMethod(writer, method_separator, method)
            method_separator=","
        writer("{}]\r\n".format(child_prefix))

    else:
        import sys
        sys.exit("Unhandled Kind '{}'".format(kind))
    #writeInlineFields(writer, ["Name","Architectures", "Type","ValueType","Value","Attrs"], "\t\t", "\r\n", constant)
    writer("{}}}\r\n".format(line_prefix))

def write(writer, api):
    writer("{\r\n")
    writer("\r\n")
    writer("\"Constants\":[\r\n")
    prefix = ""
    for constant in api["Constants"]:
        writeConstant(writer, prefix, constant)
        prefix = ","
    writer("]\r\n")
    writer("\r\n")
    writer(",\"Types\":[\r\n")
    prefix = ""
    for t in api["Types"]:
        writeType(writer, "\t", prefix, t)
        prefix = ","
    writer("]\r\n")
    writer("\r\n")
    writer(",\"Functions\":[\r\n")
    ##
    writer("]\r\n")
    writer("\r\n")
    writer(",\"UnicodeAliases\":[\r\n")
    ##
    writer("]\r\n")
    writer("\r\n")

    writer("}\r\n")


def writeInline(writer, json_val):
    if isinstance(json_val, dict):
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
    elif isinstance(json_val, str):
        writer("\"{}\"".format(json_val))
    elif isinstance(json_val, int):
        writer("{}".format(json_val))
    #if (json_val is None) or isinstance(json_val, bool) or isinstance(json_val, int):
    #    pass
    else:
        import sys
        sys.exit("{}".format(type(json_val).__name__))
    

def writeInlineFields(writer, fields, field_prefix, field_suffix, obj):
    separator = ""
    for field in fields:
        writer("{}{}\"{}\":".format(field_prefix, separator, field, field_suffix))
        writeInline(writer, obj[field])
        writer(field_suffix)
        separator = ","

def writeConstant(writer, prefix, constant):
    writer("\t{}{{\r\n".format(prefix))
    writeInlineFields(writer, ["Name","Type","ValueType","Value","Attrs"], "\t\t", "\r\n", constant)
    writer("\t}\r\n")

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
    ##
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

#!/usr/bin/env python
import sys
import xml.etree.ElementTree as ET

# create empty list for services
services = {}
assembly = {}
# create empty list for operation_lookup
operation_lookup = {}
# create empty list for error codes
error_codes = {}

def parseExceptionXML(xmlfile):
    # create element tree object
    tree = ET.parse(xmlfile)

    # get root element
    root = tree.getroot()

    # iterate httperror
    for err in root.findall('./httperror'):
        name = err.attrib['target']
        for next in err.findall('./httpresponsecode'):
            value = next.attrib['value'] or '500'
            error_codes[name] = value

def parseOperationXML(xmlfile):
    # create element tree object
    tree = ET.parse(xmlfile)

    # get root element
    root = tree.getroot()

    # iterate assembly
    assembly_name = root.attrib['assembly']

    # iterate services
    for operation in root.findall('./operation'):
        name = operation.attrib['name']
        input = None
        output = None
        errors = []
        if name != None:
            for next in operation.findall('./input'):
                input = next.attrib['target']
            for next in operation.findall('./output'):
                output = next.attrib['target']
            for next in operation.findall('./error'):
                errors.append(next.attrib['target'])
        operation_lookup[name] = {'input' : input, 'output' : output, 'errors' : errors}
        if assembly_name:
            assembly[name] = assembly_name

def parseServiceXML(xmlfile):
    # create element tree object
    tree = ET.parse(xmlfile)

    # get root element
    root = tree.getroot()

    # iterate assembly
    assembly_name = root.attrib['assembly']

    # iterate services
    for service in root.findall('./service'):
        operations = {}
        name = service.attrib['name']
        # iterate child elements of service
        for child in service:
            api = child.attrib['target']
            existing = operation_lookup[api]
            if existing:
                operations[api] = existing

        services[name] = operations
        if assembly_name:
            assembly[name] = assembly_name

def save(services, filename):
    with open(filename, 'w') as f:
        f.write('{\n')
        for i, (svc, operations) in enumerate(services.items()):
            for j, (name, operation) in enumerate(operations.items()):
                output = operation.get('output')
                target = str(assembly.get(svc, assembly.get(name, "default"))) + "." + svc + '.' + name
                f.write('"/' + name+ '":{\n')
                f.write('"post": {\n')
                f.write('  "operationId": "' + name + '",\n')
                f.write('  "requestBody": {\n')
                f.write('    "description": "' + svc + ' - ' + name + '",\n')
                f.write('    "parameters": [\n')
                f.write('    { "name": "x-amz-requestsupertrace", "in": "header", "schema": { "type": "string", "pattern": "false"}, "required": false },\n')
                f.write('    { "name": "X-Amz-Target", "in": "header", "schema": { "type": "string", "pattern": "' +
                        target + '", "example": "' + target + '", "required": false },\n')
                f.write('    { "name": "X-Requested-With", "in": "header", "schema": { "type": "string", "pattern": "XMLHTTPRequest"}, "required": false },\n')
                f.write('    { "name": "X-Amz-Date", "in": "header", "schema": { "type": "string", }, "required": false }\n')
                f.write('    ],\n')
                f.write('    "content": {\n')
                f.write('      "application/json": {\n')
                f.write('        "schema": {\n')
                f.write('          "$ref": "#/components/schemas/' + operation['input'] + '"\n')
                f.write('        }\n')
                f.write('      }\n')
                f.write('    },\n')
                f.write('    "required": true\n')
                f.write('  },\n')
                f.write('  "responses": {\n')
                f.write('    "200": {\n')
                if output:
                    f.write('      "description": "' + output + '",\n')
                f.write('      "content": {\n')
                f.write('        "application/json": {\n')
                f.write('          "schema": {\n')
                if output:
                    f.write('            "$ref": "#/components/schemas/' + output + '"\n')
                else:
                    f.write('            "type": "object"\n')
                f.write('          }\n')
                f.write('        }\n')
                f.write('      }\n')
                f.write('    },\n')
                for k, error in enumerate(operation['errors']):
                    code = error_codes.get(error) or '500'
                    f.write('    "' + code + '": {\n')
                    f.write('      "description": "' + code + ' response",\n')
                    f.write('      "content": {\n')
                    f.write('        "application/json": {\n')
                    f.write('          "schema": {\n')
                    f.write('            "type": "object"\n')
                    f.write('          }\n')
                    f.write('        }\n')
                    f.write('      }\n')
                    f.write('    }')
                    if k < len(operation['errors'])-1:
                        f.write(',')
                    f.write('\n')
                f.write('  }')
                f.write('\n')
                f.write('}\n')
                f.write('}')
                if i < len(services)-1 or j < len(operations)-1:
                    f.write(',')
                f.write('\n')
        f.write('}\n')

def main():
    for arg in sys.argv:
        if 'peration' in arg:
             parseOperationXML(arg)
        if 'xception' in arg:
             parseExceptionXML(arg)
    for arg in sys.argv:
        if 'ervice' in arg:
             parseServiceXML(arg)

    save(services, 'coral-operations.json')

if __name__ == '__main__':

    # calling main function
    main()

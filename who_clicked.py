import re
import json


with open(f'finetuned.json') as f:
    students = json.loads(f.read())

with open(f'finetuned-log.txt') as f:
    uuid_regexp = r'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
    uuids = set(re.findall(uuid_regexp, f.read()))

# actual students in the class
valid_uuids = set(students['codes'].keys())

count = 0
for uuid in uuids:
    if uuid in valid_uuids:
        count += 1
        print(uuid)

print(count)

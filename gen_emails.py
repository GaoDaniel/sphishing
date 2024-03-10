import json

with open('finetuned.json') as f:
    codes1 = json.loads(f.read())
codes1 = codes1['codes']
assert len(codes1) == 55

with open('emails.txt') as f:
    emails = eval(f.read())
assert isinstance(emails, list)

emails1 = emails[:55]
assert len(emails1) == 55

for e,c in zip(emails1, codes1):
    print(e, c)

with open('base.json') as f:
    codes2 = json.loads(f.read())

codes2 = codes2['codes']
assert len(codes2) == 54

emails2 = emails[55:]
assert len(emails2) == 54

for e,c in zip(emails2, codes2):
    print(e, c)
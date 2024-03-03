import json

# post-process the json

if __name__ == '__main__':
    with open('UGRAD.json', 'r') as f:
        threads = json.loads(f.read())
    print(len(threads))



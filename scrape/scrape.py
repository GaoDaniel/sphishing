import dotenv
import requests
import json
import os
import threading
import time

# course nums (edstem assigns a number to each course)
course_nums = {
    'DL_493G1': 50490,
    'UGRAD': 488,
    'CAREER': 855,
}

# place your jwt token in .env (key should be TOKEN)

# change this value to scrape different boards
# this is the only parameter to change
# will write to a file of the same name as BOARD
BOARD = 'CAREER'
API = f'https://us.edstem.org/api/courses/{course_nums[BOARD]}/threads'

# api for top-level threads
api = lambda limit, offset=0: API + f'?limit={limit}&offset={offset}&sort=new'

# api for individual thread (with replies)
thread_api = lambda tid: f'https://us.edstem.org/api/threads/{tid}?view=1'
threads_lk = threading.Lock()

def get_thread(url, threads):
    while True:
        thread = requests.get(url, headers=headers).json()
        if 'code' in thread and thread['code'] == 'rate_limit':
            time.sleep(3)
            continue
        break

    threads_lk.acquire()
    threads.append(thread)
    threads_lk.release()
    # print(f'total threads saved: {len(threads)}')

if __name__ == '__main__':
    dotenv.load_dotenv(os.path.join(os.path.dirname(__file__), '.env'))  
    token = os.environ.get('TOKEN')

    headers = {
        'X-Token': token
    }

    # number of threads returned
    count = 100
    offset = 0

    '''
    list of {thread: {...}, users: {...}}
    '''
    threads = list()

    while count > 0:
        data = requests.get(api(100, offset), headers=headers).json()
        if 'code' in data and data['code'] == 'rate_limit':
            # bypass ratelimiting
            time.sleep(3)
            continue

        offset += 100
        count = len(data['threads'])

        # go into each thread
        thread_urls = [thread_api(t['id']) for t in data['threads']]

        # process threads
        thds = []
        for i in range(len(thread_urls)):
            thd = threading.Thread(target=get_thread, args=(thread_urls[i], threads,))
            thd.start()
            thds.append(thd)

        for thd in thds:
            thd.join()
    
        print(f'processed {count} threads')
    
    with open(f'{BOARD}.json', 'w') as f:
        f.write(json.dumps(threads))

import json

def get_text_content(post, docs, is_admin=lambda _: True):
    if is_admin(post):
        docs.append(post['document'])
    for comment in post['comments']:
        get_text_content(comment, docs, is_admin)

# get only staff posts
def get_staff_posts(threads):
    staff_posts = []

    for t in threads:
        users = t['users']
        thread = t['thread']

        def is_admin(post):
            id = post['user_id']
            for user in users:
                if id == user['id'] and user['course_role'] == 'admin' or user['role'] == 'admin':
                    return True
            return False

        get_text_content(thread, staff_posts, is_admin)

    return staff_posts

def get_posts(threads):
    posts = []
    for t in threads:
        thread = t['thread']
        get_text_content(thread, posts)
    return posts

with open('UGRAD.json', 'r') as f:
    threads = json.loads(f.read())

staff_posts = get_staff_posts(threads)
posts = get_posts(threads)

print(len(posts))
print(len(staff_posts))
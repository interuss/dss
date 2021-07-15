import time

def example(seconds):
    print('Starting task')
    for i in range(seconds):
        print(i)
        time.sleep(1)
    print('Task completed')
    return 'task completed'

def process_auth_specs(auth_spec, config):
    print(auth_spec, config)

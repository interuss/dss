import requests
import json

ADMIN_URL = "http://localhost:8001"

def create_auth_service():
    url_path = f"{ADMIN_URL}/services"
    body = {
        "name": "auth2",
        "retries": 5,
        "protocol": "http",
        "host": "172.17.0.1",
        "port": 9096,
        "path": "/",
        "connect_timeout": 60000,
        "write_timeout": 60000,
        "read_timeout": 60000,
        "enabled": True
    }

    response = requests.post(url=url_path, json=body)
    if response.ok == False:
        print("Failed to create auth Service")
        print(response.text)
        raise Exception("Failed to create auth Service")

    return response.json()['id']

def create_auth_route(service_id):
    url_path = f"{ADMIN_URL}/services/{service_id}/routes"
    body = {
        "name": "token2",
        "methods": ["POST"],
        "paths": ["/token"],
        "strip_path": False,
        "path_handling": "v1",
        "https_redirect_status_code": 426,
        "regex_priority": 0,
        "protocols": ["http", "https"],
        "service": {
            "id": service_id
        }
    }

    response = requests.post(url=url_path, json=body)
    if response.ok == False:
        print("Failed to create auth Route")
        print(response.text)
        raise Exception("Failed to create auth Route")
    
    return response.json()['id']

def add_key_auth(route_id):
    config = {
        "key_names": ["apikey"],
        "key_in_header": True,
        "key_in_query": True,
        "key_in_body": False,
        "hide_credentials": True
    }

    body = {
        "name": "key_auth"
    }

def add_plugins(route_id):
    pass

def main():
    service_id = create_auth_service()
    create_auth_route(service_id)



if __name__ == '__main__':
    main()
import requests

class TestHarness():
    ''' A class to submit Aircraft RID State to the UTMSP test endpoint '''

    def __init__(self, test_payload):
        self.test_payload = test_payload
    
    def get_auth_token(self):
        return 'eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImp0aSI6ImEyMzBlMzRjLTNmNmUtNGU5Mi1iNjAyLTIzYjEzMmY2ODQxOSIsImlhdCI6MTYxODQxODk5NCwiZXhwIjoxNjE4NDIyNTk0fQ.O-po9I044alQuxV-EzAOgTffFXbgYyRX02XJSIy9AcI'

    def submit_test(self):

        base_url = self.test_payload['injection_url']
        
        headers = {
            'Authorization': "Bearer " + self.get_auth_token
        }

        response = requests.put(base_url, headers=headers, data=self.test_payload['injection_payload'])
        if response.status_code == 200:
            print("New test with ID %s created" % self.test_payload['injection_payload']['test_id'])
        elif response.status_code ==409:
            print("Test already with ID %s already exists" % self.test_payload['injection_payload']['test_id'])
        else: 
            print(response.json())


from faker import Faker
import string
import random
import geojson
import uuid

class OperatorFlightDataGenerator():
    ''' A class to generate fake data detailing operator name, operation name and operator location, it can be customized for locales and locations ''' 
    
    def __init__(self, language_locale):
        fake = Faker(language_locale)
        
    def generate_serial_number(self):
        return str(uuid.uuid4())
    
    def generate_registration_number(self, prefix='CHE'):
        registration_number = prefix + ''.join(random.choices(string.ascii_lowercase + string.digits, k=13))
        return registration_number
    
    def generate_operation_description(self):
        operation_description = ["Electricity Grid Inspection", "Wind farm survey", "Solar Panel Inspection", "Traffic Monitoring", "Emergency services / rescue", "Delivery operation, see more details at https://deliveryops.com/operation", "News recording, live event", "Crop spraying / Agricultural Inspection"]
        return random.choice(operation_description)
        
    def generate_operator_location(self, bbox):
        random_point = geojson.utils.generate_random("Point", boundingBox = bbox)
        operator_location = {'latitude':random_point[1], 'longitude':random_point[0]}
        return operator_location
    
    def generate_company_name(self):
        return self.fake.company()
from faker import Faker
import string
import json
import random
from itertools import cycle
# from monitoring.uss_qualifier.scd.utils import OperatorLocation, PartialOperatorDataPayload
import dataclasses

class OperatorFlightDataGenerator():
    ''' A class to generate fake data detailing operator name, operation name and operator location, it can be customized for locales and locations '''

    def __init__(self):
        self.fake = Faker()
        self.serial_number_length_code_points = {'1':1,'2':2,'3':3,'4':4,'5':5,'6':6,'7':7,'8':8,'9':9,'A':10,'B':11,'C':12,'D':13,'E':14,'F':15}
        self.serial_number_code_points = ['0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F','G','H','J','K','L','M','N','P','Q','R','S','T','U','V','W','X','Y','Z']
        self.registration_number_code_points = ['0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f','g','h','i','j','k','l','m','n','o','p','q','r','s','t','u','v','w','x','y','z']

    def generate_incorrect_serial_number(self, valid_serial_number:str) ->str:
        ''' 
        A method to modify a valid UAV serial number per ANSI/CTA-2063-A standard to one that does not conform to the standard
        
        '''
        _serial_number_length_code_points = self.serial_number_length_code_points # make a copy of the the code points
        manufacturer_code = valid_serial_number[0:4] # take out the manufacturer serial code out of the valid one
        length_code = valid_serial_number[4:5] # take out the length code out of the valid serial number         
        _serial_number_length_code_points.pop(length_code, None) # take out the length code so that we dont pick that one again (and make the serial number valid)
        dict_key, new_length_code = random.choice(list(_serial_number_length_code_points.items())) # pick a new length code
        random_serial_number = ''.join(random.choices(self.serial_number_code_points, k = new_length_code)) # generate anew 

        incorrect_serial_number =  manufacturer_code + length_code + random_serial_number

        return incorrect_serial_number


    def generate_serial_number(self) -> str:
        ''' A method to generate a random UAV serial number per ANSI/CTA-2063-A standard'''
        random.shuffle(self.serial_number_code_points)
        manufacturer_code = ''.join(self.serial_number_code_points[:4])
        dict_key, length_code = random.choice(list(self.serial_number_length_code_points.items()))
        random_serial_number = ''.join(random.choices(self.serial_number_code_points, k=length_code))

        serial_number = manufacturer_code + dict_key + random_serial_number
        return serial_number


    def generate_incorrect_registration_number(self, valid_registration_number:str) -> str:
        ''' Take a valid  Operator Registration number per the EN4709-02 standard and modify it to make it incorrect / invalid '''

        new_registration_number = valid_registration_number.split('-')[0]
        final_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(3))
        new_registration_number = new_registration_number +'-'+ final_random_string
        return new_registration_number

    def generate_registration_number(self, prefix='CHE') -> str:
        ''' A method to generate the Operator Registration number per the EN4709-02 standard '''
        def gen_checksum(raw_id):
            assert raw_id.isalnum()
            assert len(raw_id) == 15
            d = {v: k for k, v in enumerate(self.registration_number_code_points)}
            numeric_base_id = list(map(d.__getitem__, list(raw_id)))
            # Multiplication factors for each digit depending on its position
            mult_factors = cycle([2, 1])
            def partial_sum(number, mult_factor):
                """Calculate partial sum ofr a single digit."""
                quotient, remainder = divmod(number * mult_factor, 36)
                return quotient + remainder
            final_sum = sum(                
                partial_sum(int(character), mult_factor)
                for character, mult_factor in zip(numeric_base_id, mult_factors))
        
            # Calculate control number based on partial sums
            control_number = -final_sum % 36
            return self.registration_number_code_points[control_number]
        
        final_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(3))
        base_id = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(12))
        checksum = gen_checksum(base_id + final_random_string)
        reg_num = prefix + base_id + str(checksum) +'-'+ final_random_string
        return reg_num
        

    def generate_operation_description(self) -> str:
        operation_description = ["Electricity Grid Inspection", "Wind farm survey", "Solar Panel Inspection", "Traffic Monitoring", "Emergency services / rescue", "Delivery operation, see more details at https://deliveryops.com/operation", "News recording, live event", "Crop spraying / Agricultural Inspection"]
        return random.choice(operation_description)

    def generate_company_name(self) -> str:
        return self.fake.company()



if __name__ == '__main__':
    ''' This module generates a JSON that can be used to test '''
    pass
    # my_operator_flight_data_generator = OperatorFlightDataGenerator()
    # serial_number = my_operator_flight_data_generator.generate_serial_number()
    # operator_registration_number = my_operator_flight_data_generator.generate_registration_number()

    # operator_data_payload = PartialOperatorDataPayload(uas_serial_number = serial_number, operation_category='u-space', operation_mode = 'vlos',uas_class='C0', identification_technologies = 'vlos', connectivity_methods = [], endurance_minutes = [] , emergency_procedure_url = "https://uav.com/emergency", operator_id = operator_registration_number)

    # print(json.dumps(dataclasses.asdict(operator_data_payload)))

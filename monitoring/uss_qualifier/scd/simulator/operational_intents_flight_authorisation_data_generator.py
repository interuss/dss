from .utils import FlightAuthorizationData
import random

class FlightAuthorisationDataGenerator():
    ''' A class to generate data for flight authorisation per the ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664 for an UAS flight authorisation request. Reference: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e32-178-1 
    ''' 

    def __init__(self):
        '''
        This class generates a Flight Authorisation dataset, the dataset contains 11 fields at any time one of the authorisation data parameter would be incorrect this class generates a Flight Authorisation dataset 
        '''
        pass


    def generate_incorrect_serial_number(self, valid_serial_number:str) ->str:
        ''' 
        A method to modify a valid UAV serial number per ANSI/CTA-2063-A standard to one that does not conform to the standard.         
        '''
        raise NotImplementedError("Incorrect Serial Number generation not implemented")

    def generate_serial_number(self) -> str:
        ''' 
        A method to generate a random UAV serial number per ANSI/CTA-2063-A standard        
        '''
        
        raise NotImplementedError("Correct Serial Number generation not implemented")



if __name__ == '__main__':
    ''' This module generates a JSON that can be used to submit to the test interface '''
    
    my_flight_authorisation_data_generator = FlightAuthorisationDataGenerator()

    serial_number = my_flight_authorisation_data_generator.generate_serial_number()
    # TODO: Code to generate additional fields 

    make_incorrect = random.choice([0,1]) # a flag specify if one of the parameters of the flight_authorisation should be incorrect
    if make_incorrect: # if the flag is on, make the serial number incorrect        
        field_to_make_incorrect = random.choice(['uas_serial_number']) # Pick a field to make incorrect, TODO: Additional fields to be added as the generation code is impl 
        if field_to_make_incorrect == 'uas_serial_number':
            serial_number = my_flight_authorisation_data_generator.generate_incorrect_serial_number(serial_number = serial_number)

    
    flight_authorisation_data = FlightAuthorizationData(uas_serial_number = serial_number, operation_category='Open', operation_mode = 'Vlos',uas_class='C0', identification_technologies = ['ASTMNetRID'], connectivity_methods = ['cellular'], endurance_minutes = 30 , emergency_procedure_url = "https://uav.com/emergency", operator_id = 'SUSz8k1ukxjfv463-brq', uas_id= '')

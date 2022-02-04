from typing import Union
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
import random
from itertools import cycle
import string

class OperatorRegistrationNumber(str):
    """A class to represent operator registration number as expressed in the EN4709-02 standard."""
    registration_number_code_points = "0123456789abcdefghijklmnopqrstuvwxyz"

    @property
    def checksum_control(self) -> str:
        return self.split('-')[0]

    def make_invalid_by_changing_final_control_string(self,prefix='CHE') -> str:
        """A method to generate an invalid Operator Registration number by replacing the control string """        
        new_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(3))
        return OperatorRegistrationNumber(self.checksum_control +'-'+ new_random_string)

    @staticmethod
    def generate_valid(prefix='CHE') -> str:
        """A method to generate the Operator Registration number """        

        def gen_checksum(raw_id):
            assert raw_id.isalnum()
            assert len(raw_id) == 15
            
            d = {v: k for k, v in enumerate(list(OperatorRegistrationNumber.registration_number_code_points))}
            numeric_base_id = list(map(d.__getitem__, list(raw_id)))
            # Multiplication factors for each digit depending on its position
            mult_factors = cycle([2, 1])
            def partial_sum(number, mult_factor):
                #Calculate partial sum ofr a single digit.
                quotient, remainder = divmod(number * mult_factor, 36)
                return quotient + remainder
            final_sum = sum(                
                partial_sum(int(character), mult_factor)
                for character, mult_factor in zip(numeric_base_id, mult_factors))

            # Calculate control number based on partial sums
            control_number = -final_sum % 36
            
            return list(OperatorRegistrationNumber.registration_number_code_points)[control_number]

        final_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(3))
        base_id = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(12))
        checksum = gen_checksum(raw_id=base_id + final_random_string)
        
        return OperatorRegistrationNumber( prefix + base_id + str(checksum) +'-'+ final_random_string)

class SerialNumber(str):
    """Represents a serial number expressed in the ANSI/CTA-2063-A Physical Serial Number format."""
    length_code_points = "123456789ABCDEF"
    code_points = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

    @property
    def manufacturer_code(self) -> str:
        return self[0:4]

    @property
    def length_code(self) -> str:
        return self[4:5]

    def make_invalid_by_changing_payload_length(self):
        """Generates an invalid serial number similar to this serial number."""
        my_length = self.length_code
        lengths_except_mine = [c for c in SerialNumber.length_code_points if c != my_length]
        new_length_code = random.choice(lengths_except_mine)
        k = SerialNumber.length_code_points.index(new_length_code) + 1
        random_serial_number = ''.join(random.choices(SerialNumber.code_points, k=k))
        return SerialNumber(self.manufacturer_code + self.length_code + random_serial_number)

    @staticmethod
    def generate_valid():
        """Generates a valid and random UAV serial number per ANSI/CTA-2063-A."""
        manufacturer_code = ''.join(random.choices(SerialNumber.code_points, k=4))
        k = random.randrange(0, len(SerialNumber.length_code_points)) + 1
        length_code = SerialNumber.length_code_points[k - 1]
        random_serial_number = ''.join(random.choices(SerialNumber.code_points, k=k))
        return SerialNumber(manufacturer_code + length_code + random_serial_number)


class GeometryGenerationRule(ImplicitDict):
    """A class to hold configuration for developing flight paths for testing """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    """An object to hold generated flight path and the associated rule """
    geometry: Union[LineString, Polygon]    
    geometry_generation_rule: GeometryGenerationRule
   
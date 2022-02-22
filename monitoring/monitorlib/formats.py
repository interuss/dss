from __future__ import annotations

from enum import Enum
import random
import string


class OperatorRegistrationNumber(str):
    """A class to represent operator registration number as expressed in the EN4709-02 standard."""
    # TODO: Validate the behavior of this class against real data
    registration_number_code_points = "0123456789abcdefghijklmnopqrstuvwxyz"
    prefix_length = 3
    base_id_length = 12
    final_random_string_length = 3
    checksum_length = 1
    public_number_length = prefix_length + base_id_length + checksum_length
    dash_length = 1
    full_number_length = public_number_length + dash_length + final_random_string_length

    @property
    def checksum_control(self) -> str:
        return self.split('-')[0]

    @property
    def prefix(self) -> str:
        return self[0:OperatorRegistrationNumber.prefix_length]

    @property
    def base_id(self) -> str:
        return self[OperatorRegistrationNumber.prefix_length:OperatorRegistrationNumber.prefix_length + OperatorRegistrationNumber.base_id_length]

    @property
    def checksum(self) -> str:
        return self[OperatorRegistrationNumber.prefix_length + OperatorRegistrationNumber.base_id_length:][0]

    @property
    def final_random_string(self):
        return self[-OperatorRegistrationNumber.final_random_string_length:]

    @property
    def valid(self) -> bool:
        # PPPBBBBBBBBBBBBC-FFF
        # P = prefix, B = base ID, C = checksum, F = final random string
        if len(self) != OperatorRegistrationNumber.full_number_length:
            return False
        if self[OperatorRegistrationNumber.public_number_length] != '-':
            return False
        if not all(c in OperatorRegistrationNumber.registration_number_code_points for c in self.base_id):
            return False
        if not all(c in OperatorRegistrationNumber.registration_number_code_points for c in self.final_random_string):
            return False
        checksum = OperatorRegistrationNumber.generate_checksum(self.base_id, self.final_random_string)
        return self.checksum == checksum

    def make_invalid_by_changing_final_control_string(self) -> OperatorRegistrationNumber:
        """A method to generate an invalid Operator Registration number by replacing the control string """
        new_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(OperatorRegistrationNumber.final_random_string_length))
        return OperatorRegistrationNumber(self.checksum_control + '-' + new_random_string)

    @staticmethod
    def validate_prefix(prefix: str) -> None:
        if len(prefix) != OperatorRegistrationNumber.prefix_length:
            raise ValueError('Prefix of an operator registration number must be {} characters long rather than {}'.format(OperatorRegistrationNumber.prefix_length, len(prefix)))

    @staticmethod
    def validate_base_id(base_id: str) -> None:
        if len(base_id) != OperatorRegistrationNumber.base_id_length:
            raise ValueError('Base ID of an operator registration number must be {} characters long rather than {}'.format(OperatorRegistrationNumber.base_id_length, len(base_id)))
        if not all(c in OperatorRegistrationNumber.registration_number_code_points for c in base_id):
            raise ValueError('Base ID of an operator registration number must be alphanumeric')

    @staticmethod
    def validate_final_random_string(final_random_string: str) -> None:
        if len(final_random_string) != OperatorRegistrationNumber.final_random_string_length:
            raise ValueError('Final random string of an operator registration number must be {} characters long rather than {}'.format(OperatorRegistrationNumber.final_random_string_length, len(final_random_string)))
        if not all(c in OperatorRegistrationNumber.registration_number_code_points for c in final_random_string):
            raise ValueError('Final random string of an operator registration number must be alphanumeric')

    @staticmethod
    def generate_checksum(base_id: str, final_random_string: str) -> str:
        OperatorRegistrationNumber.validate_base_id(base_id)
        OperatorRegistrationNumber.validate_final_random_string(final_random_string)
        raw_id = base_id + final_random_string

        full_sum = 0
        multiplier = 2
        n = len(OperatorRegistrationNumber.registration_number_code_points)
        for c in raw_id:
            v = OperatorRegistrationNumber.registration_number_code_points.index(c)
            quotient, remainder = divmod(v * multiplier, n)
            full_sum += quotient + remainder
            multiplier = 3 - multiplier

        control_number = -full_sum % n
        return OperatorRegistrationNumber.registration_number_code_points[control_number]

    @staticmethod
    def generate_valid(prefix: str) -> OperatorRegistrationNumber:
        """Generate a random operator registration number with the specified prefix"""
        final_random_string = ''.join(random.choice(string.ascii_lowercase) for _ in range(OperatorRegistrationNumber.final_random_string_length))
        base_id = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(OperatorRegistrationNumber.base_id_length))
        return OperatorRegistrationNumber.from_components(prefix, base_id, final_random_string)

    @staticmethod
    def from_components(prefix: str, base_id: str, final_random_string: str) -> OperatorRegistrationNumber:
        """Constructs a standard operator registration number from the provided components"""
        OperatorRegistrationNumber.validate_prefix(prefix)
        OperatorRegistrationNumber.validate_base_id(base_id)
        if len(final_random_string) != OperatorRegistrationNumber.final_random_string_length:
            raise ValueError('Prefix of an operator registration number must be {} characters long rather than {}'.format(OperatorRegistrationNumber.final_random_string_length, len(final_random_string)))
        checksum = OperatorRegistrationNumber.generate_checksum(base_id, final_random_string)
        return OperatorRegistrationNumber(prefix + base_id + checksum + '-' + final_random_string)


class SerialNumber(str):
    """Represents a serial number expressed in the ANSI/CTA-2063-A Physical Serial Number format."""
    # TODO: Validate the behavior of this class against real data
    length_code_points = "123456789ABCDEF"
    code_points = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

    @property
    def manufacturer_code(self) -> str:
        return self[0:4]

    @property
    def length_code(self) -> str:
        return self[4:5]

    @property
    def manufacturer_serial_number(self) -> str:
        return self[5:]

    @property
    def valid(self) -> bool:
        if len(self) < 6:
            return False
        if not all(c in SerialNumber.code_points for c in self.manufacturer_code):
            return False
        if self.length_code not in SerialNumber.length_code_points:
            return False
        manufacturer_serial_number_length = SerialNumber.length_code_points.index(self.length_code) + 1
        if manufacturer_serial_number_length != len(self.manufacturer_serial_number):
            return False
        return True

    def make_invalid_by_changing_payload_length(self) -> SerialNumber:
        """Generates an invalid serial number similar to this serial number."""
        my_length = self.length_code
        lengths_except_mine = [c for c in SerialNumber.length_code_points if c != my_length]
        new_length_code = random.choice(lengths_except_mine)
        k = SerialNumber.length_code_points.index(new_length_code) + 1
        random_serial_number = ''.join(random.choices(SerialNumber.code_points, k=k))
        return SerialNumber(self.manufacturer_code + self.length_code + random_serial_number)

    @staticmethod
    def from_components(manufacturer_code: str, manufacturer_serial_number: str) -> SerialNumber:
        """Constructs a standard serial number from the provided components"""
        length_code = SerialNumber.length_code_points[len(manufacturer_serial_number) - 1]
        return SerialNumber(manufacturer_code + length_code + manufacturer_serial_number)

    @staticmethod
    def generate_valid() -> SerialNumber:
        """Generates a valid and random UAV serial number per ANSI/CTA-2063-A."""
        manufacturer_code = ''.join(random.choices(SerialNumber.code_points, k=4))
        k = random.randrange(0, len(SerialNumber.length_code_points)) + 1
        random_serial_number = ''.join(random.choices(SerialNumber.code_points, k=k))
        return SerialNumber.from_components(manufacturer_code, random_serial_number)

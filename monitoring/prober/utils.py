import math


ID_CHAR_SET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_'
MAX_OWNER_LENGTH = 12  # characters


def bin_to_hex(bin_string):
  return format(int(bin_string,2), "02x")


def bin_to_dec(bin_string):
  return int(bin_string, 2)


def hex_to_bin(hex_string):
  return format(int(hex_string,16), "08b")


def dec_to_bin(num):
  return format(num, "06b")


def split_by(string_val, num=8):
  """Splits a string into substrings with a length of given number."""
  return [string_val[i:i+num] for i in range(0, len(string_val), num)]


def get_ord_val(letter):
  """Encodes new ord value for ascii letters."""
  v = ID_CHAR_SET.find(letter)
  if v == -1:
    v = 63
  return v


def get_ascii_val_from_bit_value(num):
  """Decodes new ord value to ascii letters."""
  return ID_CHAR_SET[num] if num < 63 else '?'


def encode_owner(owner_name: str) -> str:
  """Encode an owner name as a 18-character hexidecimal string"""
  bits = ''
  if len(owner_name) > MAX_OWNER_LENGTH:
    string_val = owner_name[:math.ceil(MAX_OWNER_LENGTH/2)] + owner_name[-math.floor(MAX_OWNER_LENGTH/2):]
  elif len(owner_name) < MAX_OWNER_LENGTH:
    string_val = ('?' * (MAX_OWNER_LENGTH - len(owner_name))) + owner_name
  else:
    string_val = owner_name
  for letter in string_val:
    ord_val = get_ord_val(letter)
    bits += dec_to_bin(ord_val)
  return ''.join((bin_to_hex(s) for s in split_by(bits)))


def encode_resource_type_code(id_code: int) -> str:
  """Encode a number between 0 and 0xFFFF as a 4-character hexidecimal string"""
  return format(id_code, "04x")


def decode_owner(owner_id: str) -> str:
  """Decode an owner name from an 18-character hexidecimal string"""
  if len(owner_id) != 18:
    raise ValueError('Invalid owner id.')
  hex_splits = split_by(owner_id, num=2)
  bits = ''
  for h in hex_splits:
    bits += hex_to_bin(h)
  test_owner = ''
  for seq in split_by(bits, 6):
    num = bin_to_dec(seq)
    test_owner += get_ascii_val_from_bit_value(num)
  if test_owner[0] != '?':
      return test_owner[:math.ceil(MAX_OWNER_LENGTH/2)] + '..' + test_owner[-math.floor(MAX_OWNER_LENGTH/2):]
  while test_owner[0] == '?':
    test_owner = test_owner[1:]
  return test_owner


def decode_id_code(id: str) -> int:
  """Decode a number between 0 and 0xFFFF from a 4-character hex string"""
  return int(id, 16)

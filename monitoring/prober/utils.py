
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
  ord_val = ord(letter)
  if ord_val >= 48 and ord_val <= 57: # decimal numbers
    return ord_val - 48
  if ord_val >= 65 and ord_val <= 90: # capitals
    return ord_val - 65 + 10
  if ord_val >= 97 and ord_val <= 122: # small letters.
    return ord_val - 97 + 10 + 26
  if ord_val == 95:
    return 63


def get_ascii_val_from_bit_value(num):
  """Decodes new ord value to ascii letters."""
  if num >=0 and num <= 9:
    return chr(num + 48)
  if num >= 10 and num <= 35:
    return chr(num + 65 - 10)
  if num >= 36 and num <= 61:
    return chr(num + 97 - 10 - 26)
  if num == 63:
    return '_'


def encode_owner(string_val, fixed_id):
  bits = ''
  for letter in string_val:
    ord_val = get_ord_val(letter)
    bits += dec_to_bin(ord_val)
  hex_codes = ''.join((bin_to_hex(s) for s in split_by(bits)[:6]))
  fixed_code = list(fixed_id)
  curr_pos = 16
  hex_ptr = 0
  while curr_pos <= 30 and hex_ptr < len(hex_codes):
    if fixed_code[curr_pos] == '-':
      curr_pos += 1
    fixed_code[curr_pos] = hex_codes[hex_ptr]
    curr_pos += 1
    hex_ptr += 1
  return ''.join(fixed_code)


def decode_owner(owner_id):
  if len(owner_id) < 30:
    raise ValueError('Invalid owner id.')
  owner_hex_code = (owner_id[16:30]).replace('-', '')
  hex_splits = split_by(owner_hex_code, num=2)
  bits = ''
  for h in hex_splits:
    print(f'h: {h}, bits: {hex_to_bin(h)}')
    bits += hex_to_bin(h)
  test_owner = ''
  for seq in split_by(bits, 6):
    num = bin_to_dec(seq)
    test_owner += get_ascii_val_from_bit_value(num)
  return test_owner

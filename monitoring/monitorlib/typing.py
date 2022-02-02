import arrow
import datetime
from typing import get_args, get_origin, get_type_hints, Dict, Literal, Optional, Type, Union

import pytimeparse


_DICT_FIELDS = set(dir({}))
_KEY_ALL_FIELDS = '_all_fields'
_KEY_OPTIONAL_FIELDS = '_optional_fields'


class ImplicitDict(dict):
  """Base class that turns a subclass into a dict indexing attributes.

  The expected usage of this class is to declare a subclass with typed
  attributes:

    class MySubclass1(ImplicitDict):
      a: float
      b: int = 2
      c: Optional[str]
      d = 4

  All non-optional attributes must be specified when constructing an instance of
  the subclass.  Untyped attributes with a default value are considered
  optional.  In the above example, an instance of MySubclass can only be
  constructed when the values for `a` and `b` are specified in the constructor:

    MySubclass1()
      >> ValueError

    MySubclass1(a=1)
      >> ValueError

    MySubclass1(a=1, b=2)

  But, more values can be specified:

    MySubclass1(a=1, b=2, c='3', d='foo')

  The subclass allows access to all the declared attributes, but stores and
  retrieves the attributes' values in and from the underlying dict.  This means
  that the subclass will always present itself as a dict with appropriate
  entries for all declared attributes that are present.  As a result, subclasses
  are JSON-serializable by the default encoder.  Example:

    x = MySubclass1(a=1, b=2)
    x.d = 'foo'
    import json
    json.dumps(x)
      >> '{"b": 2, "d": "foo", "a": 1}'

  To deserialize JSON into an ImplicitDict subclass, use ImplicitDict.parse:

    y: MySubclass1 = ImplicitDict.parse({'b': 2, 'd': 'foo', 'a': 1}, MySubclass1)
    print(y.d)
      >> foo

  If __init__ is overridden, ImplicitDict.__init__ should be called with
  **kwargs.  For example:

    class MySubclass2(ImplicitDict):
      a: float
      b: int = 2
      c: Optional[str]
      d = 4

      def __init__(self, **kwargs):
        self.e = 5
        super(ImplicitDict, self).__init__(**kwargs)
  """

  @classmethod
  def parse(cls, source: Dict, parse_type: Type):
    if not isinstance(source, dict):
      raise ValueError('Expected to find dictionary data to populate {} object but instead found {} type'.format(parse_type.__name__, type(source).__name__))
    kwargs = {}
    hints = get_type_hints(parse_type)
    for key, value in source.items():
      if key in hints:
        # This entry has an explicit type
        kwargs[key] = _parse_value(value, hints[key])
      else:
        # This entry's type isn't specified
        kwargs[key] = value
    return parse_type(**kwargs)

  def __init__(self, previous_instance: Optional[dict]=None, **kwargs):
    super(ImplicitDict, self).__init__()
    subtype = type(self)

    if not hasattr(subtype, _KEY_ALL_FIELDS):
      # Enumerate all fields and default values defined for the subclass
      all_fields = set()
      annotations = type(self).__annotations__ if hasattr(type(self), '__annotations__') else {}
      for key in annotations:
        all_fields.add(key)

      attributes = set()
      for key in dir(self):
        if key not in _DICT_FIELDS and key[0:2] != '__' and not callable(getattr(self, key)):
          all_fields.add(key)
          attributes.add(key)

      # Identify which fields are Optional
      optional_fields = set()
      for key, field_type in annotations.items():
        generic_type = get_origin(field_type)
        if generic_type is Optional:
          optional_fields.add(key)
        elif generic_type is Union:
          generic_args = get_args(field_type)
          if len(generic_args) == 2 and generic_args[1] is type(None):
            optional_fields.add(key)
      for key in attributes:
        if key not in annotations:
          optional_fields.add(key)

      setattr(subtype, _KEY_ALL_FIELDS, all_fields)
      setattr(subtype, _KEY_OPTIONAL_FIELDS, optional_fields)
    else:
      all_fields = getattr(subtype, _KEY_ALL_FIELDS)
      optional_fields = getattr(subtype, _KEY_OPTIONAL_FIELDS)

    # Copy explicit field values passed to the constructor
    provided_values = set()
    if previous_instance:
      for key, value in previous_instance.items():
        if key in all_fields:
          self[key] = value
          provided_values.add(key)
    for key, value in kwargs.items():
      if key in all_fields:
        if value is None and key in optional_fields and key not in provided_values:
          # Don't consider an explicit null provided for an optional field as
          # actually providing a value; instead, consider it omitting a value.
          pass
        else:
          self[key] = value
          provided_values.add(key)

    # Copy default field values
    for key in all_fields:
      if key not in provided_values:
        if hasattr(type(self), key):
          self[key] = super(ImplicitDict, self).__getattribute__(key)

    # Make sure all fields without a default and not labeled Optional were provided
    for key in all_fields:
      if key not in self and key not in optional_fields:
        raise ValueError('Required field "{}" not specified in {}'.format(key, type(self).__name__))

  def __getattribute__(self, item):
    if hasattr(type(self), _KEY_ALL_FIELDS) and item in getattr(type(self), _KEY_ALL_FIELDS):
      return self[item]
    return super(ImplicitDict, self).__getattribute__(item)

  def __setattr__(self, key, value):
    if hasattr(type(self), _KEY_ALL_FIELDS):
      if key in getattr(type(self), _KEY_ALL_FIELDS):
        self[key] = value
      else:
        raise KeyError('Attribute "{}" is not defined for "{}" object'.format(key, type(self).__name__))
    else:
      super(ImplicitDict, self).__setattr__(key, value)


def _parse_value(value, value_type: Type):
  generic_type = get_origin(value_type)
  if generic_type:
    # Type is generic
    arg_types = get_args(value_type)
    if generic_type is list:
      if issubclass(arg_types[0], ImplicitDict):
        # value is a list of some kind of ImplicitDict values
        return [ImplicitDict.parse(item, arg_types[0]) for item in value]
      else:
        # value is a list of non-ImplicitDict values
        return value

    elif generic_type is Union and len(arg_types) == 2 and arg_types[1] is type(None):
      # Type is an Optional declaration
      if value is None:
        # An optional field specified explicitly as None is equivalent to
        # omitting the field's value
        return None
      else:
        return _parse_value(value, arg_types[0])

    elif generic_type is Literal and len(arg_types) == 1:
        # Type is a Literal (parsed value must match specified value)
        if value != arg_types[0]:
            raise ValueError('Value {} does not match required Literal {}'.format(value, arg_types[0]))
        return value

    else:
      raise NotImplementedError('Automatic parsing of {} type is not yet implemented'.format(value_type))

  elif issubclass(value_type, ImplicitDict):
    # value is an ImplicitDict
    return ImplicitDict.parse(value, value_type)

  else:
    # value is a non-generic type that is not an ImplicitDict
    return value_type(value) if value_type else value


class StringBasedTimeDelta(str):
  """String that only allows values which describe a timedelta."""
  def __new__(cls, value):
    if isinstance(value, str):
      dt = datetime.timedelta(seconds=pytimeparse.parse(value))
    else:
      dt = value
    str_value = str.__new__(cls, str(dt))
    str_value.timedelta = dt
    return str_value


class StringBasedDateTime(str):
  """String that only allows values which describe a datetime."""
  def __new__(cls, value):
    if isinstance(value, str):
      t = arrow.get(value).datetime
    else:
      t = value
    str_value = str.__new__(cls, arrow.get(t).to('UTC').format('YYYY-MM-DDTHH:mm:ss.SSSSSS') + 'Z')
    str_value.datetime = t
    return str_value

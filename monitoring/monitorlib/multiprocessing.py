import json
from typing import Any, Callable, Optional

import multiprocessing
import multiprocessing.shared_memory


class SynchronizedValue(object):
    """Represents a value synchronized across multiple processes.

    The shared value can be read with .value or updated in a transaction.  A
    transaction is created using `with` which returns the current value.  That
    object is mutated in the transaction, and then committed when the `with`
    block is exited.  Example:

    db = SynchronizedValue({'foo': 'bar'})
    with db as tx:
        assert isinstance(tx, dict)
        tx['foo'] = 'baz'
    print(json.dumps(db.value))
        >  {"foo":"baz"}
    """
    _lock: multiprocessing.RLock
    _shared_memory: multiprocessing.shared_memory.SharedMemory
    _encoder: Callable[[Any], bytes]
    _decoder: Callable[[bytes], Any]
    _current_value: Any

    def __init__(self, initial_value, capacity_bytes: int=10e6, encoder: Optional[Callable[[Any], bytes]]=None, decoder: Optional[Callable[[bytes], Any]]=None):
        """Creates a value synchronized across multiple processes.

        :param initial_value: Initial value to synchronize.  Must be serializable according to encoder and decoder (a dict works by default).  Must be mutatable.
        :param capacity_bytes: Maximum number of bytes required to represent this value
        :param encoder: Function that converts this value into bytes
        :param decoder: Function that converts bytes into this value
        """
        self._lock = multiprocessing.RLock()
        self._shared_memory = multiprocessing.shared_memory.SharedMemory(create=True, size=capacity_bytes)
        self._encoder = encoder if encoder is not None else lambda obj: json.dumps(obj).encode('utf-8')
        self._decoder = decoder if decoder is not None else lambda b: json.loads(b.decode('utf-8'))
        self._current_value = None
        self._set_value(initial_value)

    def _get_value(self):
        content_len = int.from_bytes(bytes(self._shared_memory.buf[0:4]), 'big')
        if content_len + 4 > self._shared_memory.size:
            raise RuntimeError('Shared memory claims to have {} bytes of content when buffer size is only {}'.format(content_len, self._shared_memory.size))
        content = bytes(self._shared_memory.buf[4:content_len + 4])
        return self._decoder(content)

    def _set_value(self, value):
        content = self._encoder(value)
        content_len = len(content)
        if content_len + 4 > self._shared_memory.size:
            raise RuntimeError('Tried to write {} bytes into a SynchronizedValue with only {} bytes of capacity'.format(content_len, self._shared_memory.size))
        self._shared_memory.buf[0:4] = content_len.to_bytes(4, 'big')
        self._shared_memory.buf[4:content_len + 4] = content

    @property
    def value(self):
        with self._lock:
            return self._get_value()

    def __enter__(self):
        self._lock.__enter__()
        self._current_value = self._get_value()
        return self._current_value

    def __exit__(self, exc_type, exc_val, exc_tb):
        try:
            if exc_type is None:
                self._set_value(self._current_value)
        finally:
            self._lock.__exit__(exc_type, exc_val, exc_tb)

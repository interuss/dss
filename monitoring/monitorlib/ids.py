import hashlib
import uuid


mac = uuid.getnode()


def make_id(code: str) -> uuid.UUID:
    """Make a pseudo-random ID that is constant on a particular machine.

    When the same `code` is specified, this ID should always be the same for a
    particular machine, but it should differ between different machines.  Its
    leading and trailing 3 bytes are hardcoded to visually indicate that it is a
    test ID.

    Args:
      code: String describing the kind of ID.  An identical code should always
            yield an identical ID on the same machine.

    Returns:
      Pseudorandom test ID in UUIDv4 format.
    """
    digest = bytearray(
        hashlib.sha1("{} {}".format(mac, code).encode("utf-8")).digest()[-16:]
    )
    digest[0] = 0
    digest[1] = 0
    digest[2] = 0xFF
    digest[-3] = 0xFF
    digest[-2] = 0
    digest[-1] = 0
    return uuid.UUID(bytes=bytes(digest), version=4)

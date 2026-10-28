package models

import (
	"uuid"

	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/stacktrace"
)

const (
	// uuidLength is the length of a raw UUID.
	uuidLength = len(uuid.UUID{})
	// uuidTextLength the length of its text form.
	uuidTextLength = 36
	// uuidsOVNLength is the length of an OVN of the form `{uuid}_{uuid}`, and packedOVNLength the length of its raw form.
	uuidsOVNLength  = 2*uuidTextLength + 1
	packedOVNLength = 2 * uuidLength
)

func OVNsToProto(ovns []OVN) []*scdpb.OVN {
	if len(ovns) == 0 {
		return nil
	}
	pbs := make([]*scdpb.OVN, len(ovns))
	for i, ovn := range ovns {
		pbs[i] = OVNToProto(ovn)
		if pbs[i] == nil {
			pbs[i] = &scdpb.OVN{}
		}
	}
	return pbs
}

func OVNsFromProto(pbs []*scdpb.OVN) ([]OVN, error) {
	if len(pbs) == 0 {
		return nil, nil
	}
	ovns := make([]OVN, len(pbs))
	for i, pb := range pbs {
		ovn, err := OVNFromProto(pb)
		if err != nil {
			return nil, err
		}
		ovns[i] = ovn
	}
	return ovns, nil
}

// OVNToProto packs the OVNs of the form `{uuid}_{uuid}` (73 chars) into their 32 raw bytes.
// Any other OVN is carried verbatim.
// An empty OVN is encoded as nil.
func OVNToProto(ovn OVN) *scdpb.OVN {
	if ovn.Empty() {
		return nil
	}
	if raw, ok := packUUIDsOVN(string(ovn)); ok {
		return &scdpb.OVN{Value: &scdpb.OVN_Uuids{Uuids: raw}}
	}
	return &scdpb.OVN{Value: &scdpb.OVN_Text{Text: string(ovn)}}
}

// OVNFromProto decodes an OVN encoded by OVNToProto.
func OVNFromProto(pb *scdpb.OVN) (OVN, error) {
	switch value := pb.GetValue().(type) {
	case nil:
		return "", nil
	case *scdpb.OVN_Uuids:
		if len(value.Uuids) != packedOVNLength {
			return "", stacktrace.NewError("invalid packed OVN length %d", len(value.Uuids))
		}
		return OVN(uuid.UUID(value.Uuids[:uuidLength]).String() + "_" + uuid.UUID(value.Uuids[uuidLength:]).String()), nil
	case *scdpb.OVN_Text:
		return OVN(value.Text), nil
	default:
		return "", stacktrace.NewError("unexpected OVN value %T", value)
	}
}

// packUUIDsOVN returns the 32 raw bytes of an OVN of the form `{uuid}_{uuid}`.
func packUUIDsOVN(s string) ([]byte, bool) {
	if len(s) != uuidsOVNLength || s[uuidTextLength] != '_' {
		return nil, false
	}
	first, second := s[:uuidTextLength], s[uuidTextLength+1:]
	a, err := uuid.Parse(first)
	if err != nil || a.String() != first {
		return nil, false
	}
	b, err := uuid.Parse(second)
	if err != nil || b.String() != second {
		return nil, false
	}
	raw := make([]byte, 0, packedOVNLength)
	raw = append(raw, a[:]...)
	raw = append(raw, b[:]...)
	return raw, true
}

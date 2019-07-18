package transport_test

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/skycoin/dmsg/cipher"

	"github.com/skycoin/skywire/pkg/transport"
)

// ExampleNewEntry shows that with different order of edges:
// - Entry.ID is the same
// - Edges() call is the same
func ExampleNewEntry() {
	pkA, _ := cipher.GenerateKeyPair()
	pkB, _ := cipher.GenerateKeyPair()

	entryAB := transport.NewEntry(pkA, pkB, "", true)
	entryBA := transport.NewEntry(pkA, pkB, "", true)

	if entryAB.ID == entryBA.ID {
		fmt.Println("entryAB.ID == entryBA.ID")
	}
	if entryAB.LocalPK() == entryBA.LocalPK() {
		fmt.Println("entryAB.LocalPK() == entryBA.LocalPK()")
	}
	if entryAB.RemotePK() == entryBA.RemotePK() {
		fmt.Println("entryAB.RemotePK() == entryBA.RemotePK()")
	}
	// Output: entryAB.ID == entryBA.ID
	// entryAB.LocalPK() == entryBA.LocalPK()
	// entryAB.RemotePK() == entryBA.RemotePK()
}

func ExampleEntry_LocalPK() {
	pkA, _ := cipher.GenerateKeyPair()
	pkB, _ := cipher.GenerateKeyPair()

	entryAB := transport.Entry{
		ID:        uuid.UUID{},
		LocalKey:  pkA,
		RemoteKey: pkB,
		Type:      "",
		Public:    true,
	}

	entryBA := transport.Entry{
		ID:        uuid.UUID{},
		LocalKey:  pkB,
		RemoteKey: pkA,
		Type:      "",
		Public:    true,
	}

	if entryAB.LocalKey != entryBA.LocalKey {
		fmt.Println("entryAB.LocalKey != entryBA.LocalKey")
	}

	if entryAB.LocalPK() == entryBA.RemotePK() {
		fmt.Println("entryAB.LocalPK() == entryBA.RemotePK()")
	}

	// Output: entryAB.LocalKey != entryBA.LocalKey
	// entryAB.LocalPK() == entryBA.RemotePK()
}

func ExampleEntry_RemotePK() {
	pkA, _ := cipher.GenerateKeyPair()
	pkB, _ := cipher.GenerateKeyPair()

	entryAB := transport.Entry{
		ID:        uuid.UUID{},
		LocalKey:  pkA,
		RemoteKey: pkB,
		Type:      "",
		Public:    true,
	}

	entryBA := transport.Entry{
		ID:        uuid.UUID{},
		LocalKey:  pkB,
		RemoteKey: pkA,
		Type:      "",
		Public:    true,
	}

	if entryAB.RemoteKey != entryBA.RemoteKey {
		fmt.Println("entryAB.RemoteKey != entryBA.RemoteKey")
	}

	if entryAB.RemotePK() == entryBA.LocalPK() {
		fmt.Println("entryAB.RemotePK() == entryBA.LocalPK()")
	}

	// Output: entryAB.RemoteKey != entryBA.RemoteKey
	// entryAB.RemotePK() == entryBA.LocalPK()
}

func ExampleEntry_SetEdges() {
	pkA, _ := cipher.GenerateKeyPair()
	pkB, _ := cipher.GenerateKeyPair()

	entryAB, entryBA := transport.Entry{}, transport.Entry{}

	entryAB.SetEdges(pkA, pkB)
	entryBA.SetEdges(pkA, pkB)

	if entryAB.LocalKey != entryBA.LocalKey {
		fmt.Println("entryAB.LocalKey != entryBA.LocalKey")
	} else {
		fmt.Println("entryAB.LocalKey == entryBA.LocalKey")
	}

	if entryAB.RemoteKey != entryBA.RemoteKey {
		fmt.Println("entryAB.RemoteKey != entryBA.RemoteKey")
	} else {
		fmt.Println("entryAB.RemoteKey != entryBA.RemoteKey")
	}

	if (entryAB.ID == entryBA.ID) && (entryAB.ID != uuid.UUID{}) {
		fmt.Println("entryAB.ID != uuid.UUID{}")
		fmt.Println("entryAB.ID == entryBA.ID")
	}

	// Output:
	// entryAB.LocalKey == entryBA.LocalKey
	// entryAB.RemoteKey != entryBA.RemoteKey
	// entryAB.ID != uuid.UUID{}
	// entryAB.ID == entryBA.ID
}

func ExampleSignedEntry_Sign() {
	pkA, skA := cipher.GenerateKeyPair()
	pkB, skB := cipher.GenerateKeyPair()

	entry := transport.NewEntry(pkA, pkB, "mock", true)
	sEntry := &transport.SignedEntry{Entry: entry}

	if sEntry.Signatures[0].Null() && sEntry.Signatures[1].Null() {
		fmt.Println("No signatures set")
	}

	if ok := sEntry.Sign(pkA, skA); !ok {
		fmt.Println("error signing with skA")
	}
	if (!sEntry.Signatures[0].Null() && sEntry.Signatures[1].Null()) ||
		(!sEntry.Signatures[1].Null() && sEntry.Signatures[0].Null()) {
		fmt.Println("One signature set")
	}

	if ok := sEntry.Sign(pkB, skB); !ok {
		fmt.Println("error signing with skB")
	}

	if !sEntry.Signatures[0].Null() && !sEntry.Signatures[1].Null() {
		fmt.Println("Both signatures set")
	} else {
		fmt.Printf("sEntry.Signatures:\n%v\n", sEntry.Signatures)
	}

	// Output: No signatures set
	// One signature set
	// Both signatures set
}

func ExampleSignedEntry_Signature() {
	pkA, skA := cipher.GenerateKeyPair()
	pkB, skB := cipher.GenerateKeyPair()

	entry := transport.NewEntry(pkA, pkB, "mock", true)
	sEntry := &transport.SignedEntry{Entry: entry}
	if ok := sEntry.Sign(pkA, skA); !ok {
		fmt.Println("Error signing sEntry with (pkA,skA)")
	}
	if ok := sEntry.Sign(pkB, skB); !ok {
		fmt.Println("Error signing sEntry with (pkB,skB)")
	}

	idxA := sEntry.Index(pkA)
	idxB := sEntry.Index(pkB)

	sigA, okA := sEntry.Signature(pkA)
	sigB, okB := sEntry.Signature(pkB)

	if okA && sigA == sEntry.Signatures[idxA] {
		fmt.Println("SignatureA got")
	}

	if okB && (sigB == sEntry.Signatures[idxB]) {
		fmt.Println("SignatureB got")
	}

	// Incorrect case
	pkC, _ := cipher.GenerateKeyPair()
	if _, ok := sEntry.Signature(pkC); !ok {
		fmt.Printf("SignatureC got error: invalid pubkey")
	}

	//
	// Output: SignatureA got
	// SignatureB got
	// SignatureC got error: invalid pubkey
}

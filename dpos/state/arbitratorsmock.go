package state

import (
	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types"
	"github.com/elastos/Elastos.ELA/core/types/payload"
	"github.com/elastos/Elastos.ELA/dpos/p2p"
)

func NewArbitratorsMock(arbitersByte [][]byte, changeCount, majorityCount int) *ArbitratorsMock {
	return &ArbitratorsMock{
		CurrentArbitrators:          arbitersByte,
		CurrentCandidates:           make([][]byte, 0),
		NextArbitrators:             make([][]byte, 0),
		NextCandidates:              make([][]byte, 0),
		CurrentOwnerProgramHashes:   make([]*common.Uint168, 0),
		CandidateOwnerProgramHashes: make([]*common.Uint168, 0),
		OwnerVotesInRound:           make(map[common.Uint168]common.Fixed64),
		TotalVotesInRound:           0,
		DutyChangedCount:            0,
		MajorityCount:               majorityCount,
	}
}

//mock object of arbitrators
type ArbitratorsMock struct {
	CurrentArbitrators          [][]byte
	CurrentCandidates           [][]byte
	NextArbitrators             [][]byte
	NextCandidates              [][]byte
	CurrentOwnerProgramHashes   []*common.Uint168
	CandidateOwnerProgramHashes []*common.Uint168
	OwnerVotesInRound           map[common.Uint168]common.Fixed64
	TotalVotesInRound           common.Fixed64
	DutyChangedCount            int
	MajorityCount               int
}

func (a *ArbitratorsMock) GetDutyIndexByHeight(height uint32) int {
	panic("implement me")
}

func (a *ArbitratorsMock) GetDutyIndex() int {
	panic("implement me")
}

func (a *ArbitratorsMock) ProcessSpecialTxPayload(p types.Payload, height uint32) error {
	panic("implement me")
}

func (a *ArbitratorsMock) ProcessBlock(block *types.Block, confirm *payload.Confirm) {
	panic("implement me")
}

func (a *ArbitratorsMock) RollbackTo(height uint32) error {
	panic("implement me")
}

func (a *ArbitratorsMock) GetNeedConnectArbiters(height uint32) map[string]*p2p.PeerAddr {
	panic("implement me")
}

func (a *ArbitratorsMock) IsArbitrator(pk []byte) bool {
	panic("implement me")
}

func (a *ArbitratorsMock) IsCRCArbitrator(pk []byte) bool {
	panic("implement me")
}

func (a *ArbitratorsMock) GetLastConfirmedBlockTimeStamp() uint32 {
	panic("implement me")
}

func (a *ArbitratorsMock) TryEnterEmergency(blockTime uint32) bool {
	panic("implement me")
}

func (a *ArbitratorsMock) GetCRCProducer(publicKey []byte) *Producer {
	panic("implement me")
}

func (a *ArbitratorsMock) GetCRCArbitrators() map[string]*Producer {
	panic("implement me")
}

func (a *ArbitratorsMock) IsCRCArbitratorProgramHash(hash *common.Uint168) bool {
	return false
}

func (a *ArbitratorsMock) IsCRCArbitratorNodePublicKey(nodePublicKeyHex string) bool {
	return false
}

func (a *ArbitratorsMock) GetArbitersCount() int {
	return len(a.CurrentArbitrators)
}

func (a *ArbitratorsMock) GetArbitersMajorityCount() int {
	return a.MajorityCount
}

func (a *ArbitratorsMock) GetDutyChangeCount() int {
	return a.DutyChangedCount
}

func (a *ArbitratorsMock) SetDutyChangeCount(count int) {
	a.DutyChangedCount = count
}

func (a *ArbitratorsMock) GetArbitrators() [][]byte {
	return a.CurrentArbitrators
}

func (a *ArbitratorsMock) GetNormalArbitrators() ([][]byte, error) {
	return a.CurrentArbitrators, nil
}

func (a *ArbitratorsMock) GetCandidates() [][]byte {
	return a.CurrentCandidates
}

func (a *ArbitratorsMock) GetNextArbitrators() [][]byte {
	return a.NextArbitrators
}

func (a *ArbitratorsMock) GetNextCandidates() [][]byte {
	return a.NextCandidates
}

func (a *ArbitratorsMock) GetDutyChangedCount() int {
	return a.DutyChangedCount
}

func (a *ArbitratorsMock) SetDutyChangedCount(count int) {
	a.DutyChangedCount = count
}

func (a *ArbitratorsMock) SetArbitrators(ar [][]byte) {
	a.CurrentArbitrators = ar
}

func (a *ArbitratorsMock) SetCandidates(ca [][]byte) {
	a.CurrentCandidates = ca
}

func (a *ArbitratorsMock) SetNextArbitrators(ar [][]byte) {
	a.NextArbitrators = ar
}

func (a *ArbitratorsMock) SetNextCandidates(ca [][]byte) {
	a.NextCandidates = ca
}

func (a *ArbitratorsMock) GetCurrentOwnerProgramHashes() []*common.Uint168 {
	result := a.CurrentOwnerProgramHashes

	return result
}

func (a *ArbitratorsMock) GetCandidateOwnerProgramHashes() []*common.Uint168 {
	result := a.CandidateOwnerProgramHashes

	return result
}

func (a *ArbitratorsMock) GetOwnerVotes(programHash *common.Uint168) common.Fixed64 {
	result := a.OwnerVotesInRound[*programHash]

	return result
}

func (a *ArbitratorsMock) GetTotalVotesInRound() common.Fixed64 {
	result := a.TotalVotesInRound

	return result
}

func (a *ArbitratorsMock) GetOnDutyArbitrator() []byte {
	return a.GetNextOnDutyArbitrator(0)
}

func (a *ArbitratorsMock) GetNextOnDutyArbitrator(offset uint32) []byte {
	if len(a.CurrentArbitrators) == 0 {
		return nil
	}
	index := (a.DutyChangedCount + int(offset)) % len(a.CurrentArbitrators)
	return a.CurrentArbitrators[index]
}

func (a *ArbitratorsMock) HasArbitersMajorityCount(num int) bool {
	return num > a.MajorityCount
}

func (a *ArbitratorsMock) HasArbitersMinorityCount(num int) bool {
	return num >= len(a.CurrentArbitrators)-a.MajorityCount
}

func (a *ArbitratorsMock) DumpInfo() {
	panic("implement me")
}

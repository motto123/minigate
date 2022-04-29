package session

import "sync"

type SyncClientMap struct {
	sync.RWMutex
	sessionM   map[uint64]Session
	uidAndSidM map[int64]uint64
}

func NewSyncSessionMap() *SyncClientMap {
	return &SyncClientMap{
		RWMutex:    sync.RWMutex{},
		sessionM:   make(map[uint64]Session),
		uidAndSidM: make(map[int64]uint64),
	}
}

func (s *SyncClientMap) Store(key uint64, val Session) {
	s.Lock()
	defer s.Unlock()
	s.sessionM[key] = val
}

func (s *SyncClientMap) Delete(key uint64) {
	s.Lock()
	defer s.Unlock()

	session2 := s.sessionM[key]
	delete(s.uidAndSidM, session2.GetUid())
	delete(s.sessionM, key)
}

func (s *SyncClientMap) Load(key uint64) Session {
	s.RLock()
	defer s.RUnlock()
	return s.sessionM[key]
}

func (s *SyncClientMap) Range(f func(key uint64, value Session) bool) {
	s.RLock()
	defer s.RUnlock()
	for k, v := range s.sessionM {
		if v == nil {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

func (s *SyncClientMap) StoreWithUid(uid int64, val Session) {
	s.Lock()
	defer s.Unlock()
	s.uidAndSidM[uid] = val.GetId()
}

func (s *SyncClientMap) StoreWithUidAndSid(uid int64, sid uint64) {
	s.Lock()
	defer s.Unlock()
	s.uidAndSidM[uid] = sid
}

func (s *SyncClientMap) DeleteWithUid(uid int64) {
	s.Lock()
	defer s.Unlock()
	sid := s.uidAndSidM[uid]
	delete(s.sessionM, sid)
	delete(s.uidAndSidM, uid)
}

func (s *SyncClientMap) LoadWithUid(uid int64) Session {
	s.RLock()
	defer s.RUnlock()
	sid := s.uidAndSidM[uid]
	return s.sessionM[sid]
}

// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"context"
	"sync"

	"github.com/RTradeLtd/grpc/lensv2"
	"google.golang.org/grpc"
)

type FakeLensV2Client struct {
	IndexStub        func(context.Context, *lensv2.IndexReq, ...grpc.CallOption) (*lensv2.IndexResp, error)
	indexMutex       sync.RWMutex
	indexArgsForCall []struct {
		arg1 context.Context
		arg2 *lensv2.IndexReq
		arg3 []grpc.CallOption
	}
	indexReturns struct {
		result1 *lensv2.IndexResp
		result2 error
	}
	indexReturnsOnCall map[int]struct {
		result1 *lensv2.IndexResp
		result2 error
	}
	RemoveStub        func(context.Context, *lensv2.RemoveReq, ...grpc.CallOption) (*lensv2.RemoveResp, error)
	removeMutex       sync.RWMutex
	removeArgsForCall []struct {
		arg1 context.Context
		arg2 *lensv2.RemoveReq
		arg3 []grpc.CallOption
	}
	removeReturns struct {
		result1 *lensv2.RemoveResp
		result2 error
	}
	removeReturnsOnCall map[int]struct {
		result1 *lensv2.RemoveResp
		result2 error
	}
	SearchStub        func(context.Context, *lensv2.SearchReq, ...grpc.CallOption) (*lensv2.SearchResp, error)
	searchMutex       sync.RWMutex
	searchArgsForCall []struct {
		arg1 context.Context
		arg2 *lensv2.SearchReq
		arg3 []grpc.CallOption
	}
	searchReturns struct {
		result1 *lensv2.SearchResp
		result2 error
	}
	searchReturnsOnCall map[int]struct {
		result1 *lensv2.SearchResp
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLensV2Client) Index(arg1 context.Context, arg2 *lensv2.IndexReq, arg3 ...grpc.CallOption) (*lensv2.IndexResp, error) {
	fake.indexMutex.Lock()
	ret, specificReturn := fake.indexReturnsOnCall[len(fake.indexArgsForCall)]
	fake.indexArgsForCall = append(fake.indexArgsForCall, struct {
		arg1 context.Context
		arg2 *lensv2.IndexReq
		arg3 []grpc.CallOption
	}{arg1, arg2, arg3})
	fake.recordInvocation("Index", []interface{}{arg1, arg2, arg3})
	fake.indexMutex.Unlock()
	if fake.IndexStub != nil {
		return fake.IndexStub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.indexReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeLensV2Client) IndexCallCount() int {
	fake.indexMutex.RLock()
	defer fake.indexMutex.RUnlock()
	return len(fake.indexArgsForCall)
}

func (fake *FakeLensV2Client) IndexCalls(stub func(context.Context, *lensv2.IndexReq, ...grpc.CallOption) (*lensv2.IndexResp, error)) {
	fake.indexMutex.Lock()
	defer fake.indexMutex.Unlock()
	fake.IndexStub = stub
}

func (fake *FakeLensV2Client) IndexArgsForCall(i int) (context.Context, *lensv2.IndexReq, []grpc.CallOption) {
	fake.indexMutex.RLock()
	defer fake.indexMutex.RUnlock()
	argsForCall := fake.indexArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLensV2Client) IndexReturns(result1 *lensv2.IndexResp, result2 error) {
	fake.indexMutex.Lock()
	defer fake.indexMutex.Unlock()
	fake.IndexStub = nil
	fake.indexReturns = struct {
		result1 *lensv2.IndexResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) IndexReturnsOnCall(i int, result1 *lensv2.IndexResp, result2 error) {
	fake.indexMutex.Lock()
	defer fake.indexMutex.Unlock()
	fake.IndexStub = nil
	if fake.indexReturnsOnCall == nil {
		fake.indexReturnsOnCall = make(map[int]struct {
			result1 *lensv2.IndexResp
			result2 error
		})
	}
	fake.indexReturnsOnCall[i] = struct {
		result1 *lensv2.IndexResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) Remove(arg1 context.Context, arg2 *lensv2.RemoveReq, arg3 ...grpc.CallOption) (*lensv2.RemoveResp, error) {
	fake.removeMutex.Lock()
	ret, specificReturn := fake.removeReturnsOnCall[len(fake.removeArgsForCall)]
	fake.removeArgsForCall = append(fake.removeArgsForCall, struct {
		arg1 context.Context
		arg2 *lensv2.RemoveReq
		arg3 []grpc.CallOption
	}{arg1, arg2, arg3})
	fake.recordInvocation("Remove", []interface{}{arg1, arg2, arg3})
	fake.removeMutex.Unlock()
	if fake.RemoveStub != nil {
		return fake.RemoveStub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.removeReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeLensV2Client) RemoveCallCount() int {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return len(fake.removeArgsForCall)
}

func (fake *FakeLensV2Client) RemoveCalls(stub func(context.Context, *lensv2.RemoveReq, ...grpc.CallOption) (*lensv2.RemoveResp, error)) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = stub
}

func (fake *FakeLensV2Client) RemoveArgsForCall(i int) (context.Context, *lensv2.RemoveReq, []grpc.CallOption) {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	argsForCall := fake.removeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLensV2Client) RemoveReturns(result1 *lensv2.RemoveResp, result2 error) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = nil
	fake.removeReturns = struct {
		result1 *lensv2.RemoveResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) RemoveReturnsOnCall(i int, result1 *lensv2.RemoveResp, result2 error) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = nil
	if fake.removeReturnsOnCall == nil {
		fake.removeReturnsOnCall = make(map[int]struct {
			result1 *lensv2.RemoveResp
			result2 error
		})
	}
	fake.removeReturnsOnCall[i] = struct {
		result1 *lensv2.RemoveResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) Search(arg1 context.Context, arg2 *lensv2.SearchReq, arg3 ...grpc.CallOption) (*lensv2.SearchResp, error) {
	fake.searchMutex.Lock()
	ret, specificReturn := fake.searchReturnsOnCall[len(fake.searchArgsForCall)]
	fake.searchArgsForCall = append(fake.searchArgsForCall, struct {
		arg1 context.Context
		arg2 *lensv2.SearchReq
		arg3 []grpc.CallOption
	}{arg1, arg2, arg3})
	fake.recordInvocation("Search", []interface{}{arg1, arg2, arg3})
	fake.searchMutex.Unlock()
	if fake.SearchStub != nil {
		return fake.SearchStub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.searchReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeLensV2Client) SearchCallCount() int {
	fake.searchMutex.RLock()
	defer fake.searchMutex.RUnlock()
	return len(fake.searchArgsForCall)
}

func (fake *FakeLensV2Client) SearchCalls(stub func(context.Context, *lensv2.SearchReq, ...grpc.CallOption) (*lensv2.SearchResp, error)) {
	fake.searchMutex.Lock()
	defer fake.searchMutex.Unlock()
	fake.SearchStub = stub
}

func (fake *FakeLensV2Client) SearchArgsForCall(i int) (context.Context, *lensv2.SearchReq, []grpc.CallOption) {
	fake.searchMutex.RLock()
	defer fake.searchMutex.RUnlock()
	argsForCall := fake.searchArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLensV2Client) SearchReturns(result1 *lensv2.SearchResp, result2 error) {
	fake.searchMutex.Lock()
	defer fake.searchMutex.Unlock()
	fake.SearchStub = nil
	fake.searchReturns = struct {
		result1 *lensv2.SearchResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) SearchReturnsOnCall(i int, result1 *lensv2.SearchResp, result2 error) {
	fake.searchMutex.Lock()
	defer fake.searchMutex.Unlock()
	fake.SearchStub = nil
	if fake.searchReturnsOnCall == nil {
		fake.searchReturnsOnCall = make(map[int]struct {
			result1 *lensv2.SearchResp
			result2 error
		})
	}
	fake.searchReturnsOnCall[i] = struct {
		result1 *lensv2.SearchResp
		result2 error
	}{result1, result2}
}

func (fake *FakeLensV2Client) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.indexMutex.RLock()
	defer fake.indexMutex.RUnlock()
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	fake.searchMutex.RLock()
	defer fake.searchMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLensV2Client) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ lensv2.LensV2Client = new(FakeLensV2Client)

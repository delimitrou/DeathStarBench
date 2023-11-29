package consul

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/jpillora/backoff"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

// init function needs for  auto-register in resolvers registry
func init() {
	resolver.Register(&builder{})
}

// resolvr implements resolver.Resolver from the gRPC package.
// It watches for endpoints changes and pushes them to the underlying gRPC connection.
type resolvr struct {
	cancelFunc context.CancelFunc
}

// ResolveNow will be skipped due unnecessary in this case
func (r *resolvr) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (r *resolvr) Close() {
	r.cancelFunc()
}

//go:generate ./bin/moq -out mocks_test.go . servicer
type servicer interface {
	Service(string, string, bool, *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)
}

func watchConsulService(ctx context.Context, s servicer, tgt target, out chan<- []string) {
	res := make(chan []string)
	quit := make(chan struct{})
	bck := &backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    10 * time.Millisecond,
		Max:    tgt.MaxBackoff,
	}
	go func() {
		var lastIndex uint64
		for {
			ss, meta, err := s.Service(
				tgt.Service,
				tgt.Tag,
				tgt.Healthy,
				&api.QueryOptions{
					WaitIndex:         lastIndex,
					Near:              tgt.Near,
					WaitTime:          tgt.Wait,
					Datacenter:        tgt.Dc,
					AllowStale:        tgt.AllowStale,
					RequireConsistent: tgt.RequireConsistent,
				},
			)
			if err != nil {
				// No need to continue if the context is done/cancelled.
				// We check that here directly because the check for the closed quit channel
				// at the end of the loop is not reached when calling continue here.
				select {
				case <-quit:
					return
				default:
					grpclog.Errorf("[Consul resolver] Couldn't fetch endpoints. target={%s}; error={%v}", tgt.String(), err)
					time.Sleep(bck.Duration())
					continue
				}
			}
			bck.Reset()
			lastIndex = meta.LastIndex
			grpclog.Infof("[Consul resolver] %d endpoints fetched in(+wait) %s for target={%s}",
				len(ss),
				meta.RequestTime,
				tgt.String(),
			)

			ee := make([]string, 0, len(ss))
			for _, s := range ss {
				address := s.Service.Address
				if s.Service.Address == "" {
					address = s.Node.Address
				}
				ee = append(ee, fmt.Sprintf("%s:%d", address, s.Service.Port))
			}

			if tgt.Limit != 0 && len(ee) > tgt.Limit {
				ee = ee[:tgt.Limit]
			}
			select {
			case res <- ee:
				continue
			case <-quit:
				return
			}
		}
	}()

	for {
		// If in the below select both channels have values that can be read,
		// Go picks one pseudo-randomly.
		// But when the context is canceled we want to act upon it immediately.
		if ctx.Err() != nil {
			// Close quit so the goroutine returns and doesn't leak.
			// Do NOT close res because that can lead to panics in the goroutine.
			// res will be garbage collected at some point.
			close(quit)
			return
		}
		select {
		case ee := <-res:
			out <- ee
		case <-ctx.Done():
			close(quit)
			return
		}
	}
}

func populateEndpoints(ctx context.Context, clientConn resolver.ClientConn, input <-chan []string) {
	for {
		select {
		case cc := <-input:
			connsSet := make(map[string]struct{}, len(cc))
			for _, c := range cc {
				connsSet[c] = struct{}{}
			}
			conns := make([]resolver.Address, 0, len(connsSet))
			for c := range connsSet {
				conns = append(conns, resolver.Address{Addr: c})
			}
			sort.Sort(byAddressString(conns)) // Don't replace the same address list in the balancer
			err := clientConn.UpdateState(resolver.State{Addresses: conns})
			if err != nil {
				grpclog.Errorf("[Consul resolver] Couldn't update client connection. error={%v}", err)
				continue
			}
		case <-ctx.Done():
			grpclog.Info("[Consul resolver] Watch has been finished")
			return
		}
	}
}

// byAddressString sorts resolver.Address by Address Field  sorting in increasing order.
type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

package main

import _ "fmt"
import . "github.com/go-fed/activity/pub"
import _ "github.com/go-fed/activity/streams"
import "github.com/go-fed/activity/streams/vocab"
import "net/http"
import "context"
import "sync"
import "net/url"
import "errors"
import "time"

func main() {
	s := &myService{}
	db := &myDB{
		content:  &sync.Map{},
		locks:    &sync.Map{},
		hostname: "localhost",
	}
	//actor := NewFederatingActor( /* CommonBehavior */ s,
	_ = NewFederatingActor( /* CommonBehavior */ s,
		/* FederatingProtocol */ s,
		/* Database */ db,
		/* Clock */ s)
}

// stub it out!
//func NewFederatingActor(c CommonBehavior,
//	s2s FederatingProtocol,
//	db Database,
//	clock Clock) FederatingActor

type myService struct{}

func (*myService) AuthenticateGetInbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) AuthenticateGetOutbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) GetOutbox(c context.Context,
	r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}

func (*myService) NewTransport(c context.Context,
	actorBoxIRI *url.URL,
	gofedAgent string) (t Transport, err error) {
	// TODO
	return
}

func (*myService) PostInboxRequestBodyHook(c context.Context,
	r *http.Request,
	activity Activity) (context.Context, error) {
	// TODO
	return nil, nil
}

func (*myService) AuthenticatePostInbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) Blocked(c context.Context,
	actorIRIs []*url.URL) (blocked bool, err error) {
	// TODO
	return
}

func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	// TODO
	return
}

func (*myService) DefaultCallback(c context.Context,
	activity Activity) error {
	// TODO
	return nil
}

func (*myService) MaxInboxForwardingRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (*myService) MaxDeliveryRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (*myService) FilterForwarding(c context.Context,
	potentialRecipients []*url.URL,
	a Activity) (filteredRecipients []*url.URL, err error) {
	// TODO
	return
}

func (*myService) GetInbox(c context.Context,
	r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}

type myDB struct {
	// The content of our app, keyed by ActivityPub ID.
	content *sync.Map
	// Enables mutations. A sync.Mutex per ActivityPub ID.
	locks *sync.Map
	// The host domain of our service, for detecting ownership.
	hostname string
}

// Our content map will store this data.
type content struct {
	// The payload of the data: vocab.Type is any type understood by Go-Fed.
	data vocab.Type
	// If true, belongs to our local user and not a federated peer. This is
	// recommended for a solution that just indiscriminately puts everything
	// into a single "table", like this in-memory solution.
	isLocal bool
}

func (m *myDB) Lock(c context.Context,
	id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	mu := &sync.Mutex{}
	mu.Lock() // Optimistically lock if we do store it.
	i, loaded := m.locks.LoadOrStore(id.String(), mu)
	if loaded {
		mu = i.(*sync.Mutex)
		mu.Lock()
	}
	return nil
}

func (m *myDB) Unlock(c context.Context,
	id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.

	i, ok := m.locks.Load(id.String())
	if !ok {
		return errors.New("Missing an id in Unlock")
	}
	mu := i.(*sync.Mutex)
	mu.Unlock()
	return nil
}

func (m *myDB) Owns(c context.Context,
	id *url.URL) (owns bool, err error) {
	// Owns just determines if the ActivityPub id is owned by this server.
	// In a real implementation, consider something far more robust than
	// this string comparison.
	return id.Host == m.hostname, nil
}

func (m *myDB) Exists(c context.Context,
	id *url.URL) (exists bool, err error) {
	// Do we have this `id`?
	_, exists = m.content.Load(id.String())
	return
}

/*
func (m *myDB) Get(c context.Context,
	id *url.URL) (value vocab.Type, err error) {
	// Our goal is to return what we have at that `id`. Returns an error if
	// not found.
	iCon, exists = m.content.Load(id.String())
	if !exists {
		err = errors.New("Get failed")
		return
	}
	// Extract the data from our `content` type.
	con := iCon.(*content)
	return con.data
}

func (m *myDB) Create(c context.Context,
	asType vocab.Type) error {
	// Create a payload in our in-memory map. The thing could be a local or
	// a federated peer's data. We can re-use the `Owns` call to set the
	// metadata on our `content`.
	id, err := GetId(asType)
	if err != nil {
		return err
	}
	owns, err := m.Owns(id)
	if err != nil {
		return err
	}
	con = &content{
		data:    asType,
		isLocal: owns,
	}
	m.content.Store(id.String(), con)
	return nil
}

func (m *myDB) Update(c context.Context,
	asType vocab.Type) error {
	// Replace a payload in our in-memory map. The thing could be a local or
	// a federated peer's data. Since we are using a map and not a solution
	// like SQL, we can simply do what `Create` does: overwrite it.
	//
	// Note that an actor's followers, following, and liked collections are
	// never Created, only Updated.
	return m.Create(c, asType)
}

func (m *myDB) Delete(c context.Context,
	id *url.URL) error {
	// Remove a payload in our in-memory map.
	m.Delete(id.String())
	return nil
}

func (m *myDB) InboxContains(c context.Context,
	inbox,
	id *url.URL) (contains bool, err error) {
	// Our goal is to see if the `inbox`, which is an OrderedCollection,
	// contains an element in its `ordered_items` property that has a
	// matching `id`
	contains = false
	var oc vocab.ActivityStreamsOrderedCollection
	// getOrderedCollection is a helper method to fetch an
	// OrderedCollection. It is not implemented in this tutorial, and uses
	// the map m.content to do the lookup.
	oc, err = m.getOrderedCollection(inbox)
	if err != nil {
		return
	}
	// Next, we use the ActivityStreams vocabulary to obtain the
	// ordered_items property of the OrderedCollection type.
	oi := oc.GetActivityStreamsOrderedItems()
	// Properties may be nil, if non-existent!
	if oi == nil {
		return
	}
	// Finally, loop through each item in the ordered_items property and see
	// if the element's id matches the desired id.
	for iter := oi.Begin(); iter != oi.End(); iter = iter.Next() {
		var iterId *url.URL
		iterId, err = ToId(iter)
		if err != nil {
			return
		}
		if iterId.String() == id.String() {
			contains = true
			return
		}
	}
	return
}

func (m *myDB) GetInbox(c context.Context,
	inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// The goal here is to fetch an inbox at the specified IRI.

	// getOrderedCollectionPage is a helper method to fetch an
	// OrderedCollectionPage. It is not implemented in this tutorial, and
	// uses the map m.content to do the lookup and any conversions if
	// needed. The database can get fancy and use query parameters in the
	// `inboxIRI` to paginate appropriately.
	return m.getOrderedCollectionPage(inboxIRI)
}

func (m *myDB) SetInbox(c context.Context,
	inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// The goal here is to set an inbox at the specified IRI, with any
	// changes to the page made persistent. Since the inbox has been Locked,
	// it is OK to assume that no other concurrent goroutine has changed the
	// inbox in the meantime.

	// getOrderedCollection is a helper method to fetch an
	// OrderedCollection. It is not implemented in this tutorial, and
	// uses the map m.content to do the lookup.
	storedInbox, err := m.getOrderedCollection(inboxIRI)
	if err != nil {
		return err
	}
	// applyDiffOrderedCollection is a helper method to apply changes due
	// to an edited OrderedCollectionPage. Implementation is left as an
	// exercise for the reader.
	updatedInbox := m.applyDiffOrderedCollection(storedInbox, inbox)

	// saveToContent is a helper method to save an
	// ActivityStream type. Implementation is left as an exercise for the
	// reader.
	return m.saveToContent(updatedInbox)
}

func (m *myDB) GetOutbox(c context.Context,
	inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// Similar to `GetInbox`, but for the outbox. See `GetInbox`.
}

func (m *myDB) SetOutbox(c context.Context,
	inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// Similar to `SetInbox`, but for the outbox. See `SetInbox`.
}

func (m *myDB) ActorForOutbox(c context.Context,
	outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	// Given the `outboxIRI`, determine the IRI of the actor that owns
	// that outbox. Will only be used for actors on this local server.
	// Implementation left as an exercise to the reader.
}
*/

func (m *myDB) ActorForInbox(c context.Context,
	inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	// Given the `inboxIRI`, determine the IRI of the actor that owns
	// that inbox. Will only be used for actors on this local server.
	// Implementation left as an exercise to the reader.
	return
}

/*
func (m *myDB) OutboxForInbox(c context.Context,

		inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
		// Given the `inboxIRI`, determine the IRI of the outbox owned
		// by the same actor that owns the inbox. Will only be used for actors
		// on this local server. Implementation left as an exercise to the
		// reader.
	}

func (m *myDB) NewID(c context.Context,

		t vocab.Type) (id *url.URL, err error) {
		// Generate a new `id` for the ActivityStreams object `t`.

		// You can be fancy and put different types authored by different folks
		// along different paths. Or just generate a GUID. Implementation here
		// is left as an exercise for the reader.
	}

func (m *myDB) Followers(c context.Context,

		actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
		// Get the followers collection from the actor with `actorIRI`.

		// getPerson is a helper method that returns an actor on this server
		// with a Person ActivityStreams type. It is not implemented in this tutorial.
		var person vocab.ActivityStreamsPerson
		person, err = m.getPerson(actorIRI)
		if err != nil {
			return
		}
		// Let's get their followers property, ensure it exists, and then
		// fetch it with a familiar helper method.
		f := person.GetActivityStreamsFollowers()
		if f == nil {
			err = errors.New("no followers collection")
			return
		}
		// Note: at this point f is not the OrderedCollection itself yet. It is
		// an opaque box (it could be an IRI, an OrderedCollection, or something
		// extending an OrderedCollection).
		followersId, err := ToId(f)
		if err != nil {
			return
		}
		return m.getOrderedCollection(followersId)
	}

func (m *myDB) Following(c context.Context,

		actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
		// Get the following collection from the actor with `actorIRI`.

		// Implementation is similar to `Followers`. See `Followers`.
	}

func (m *myDB) Liked(c context.Context,

		actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
		// Get the liked collection from the actor with `actorIRI`.

		// Implementation is similar to `Followers`. See `Followers`.
	}
*/
func (*myService) Now() time.Time {
	return time.Now()
}

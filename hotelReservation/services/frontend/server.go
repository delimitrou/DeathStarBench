package frontend

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/dialer"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	attractions "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/attractions/proto"
	profile "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/profile/proto"
	recommendation "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/recommendation/proto"
	reservation "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/reservation/proto"
	review "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/review/proto"
	search "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/search/proto"
	user "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/user/proto"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tls"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	//go:embed static/*
	content embed.FS
)

// Server implements frontend service
type Server struct {
	searchClient         search.SearchClient
	profileClient        profile.ProfileClient
	recommendationClient recommendation.RecommendationClient
	userClient           user.UserClient
	reviewClient         review.ReviewClient
	attractionsClient    attractions.AttractionsClient
	reservationClient    reservation.ReservationClient

	KnativeDns string
	IpAddr     string
	ConsulAddr string
	Port       int
	Tracer     opentracing.Tracer
	Registry   *registry.Client
}

// Run the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("Server port must be set")
	}

	log.Info().Msg("Loading static content...")
	staticContent, err := fs.Sub(content, "static")
	if err != nil {
		return err
	}

	log.Info().Msg("Initializing gRPC clients...")
	if err := s.initSearchClient("srv-search"); err != nil {
		return err
	}

	if err := s.initProfileClient("srv-profile"); err != nil {
		return err
	}

	if err := s.initRecommendationClient("srv-recommendation"); err != nil {
		return err
	}

	if err := s.initUserClient("srv-user"); err != nil {
		return err
	}

	if err := s.initReservation("srv-reservation"); err != nil {
		return err
	}

	if err := s.initReviewClient("srv-review"); err != nil {
		return err
	}

	if err := s.initAttractionsClient("srv-attractions"); err != nil {
		return err
	}

	log.Info().Msg("Successful")

	log.Trace().Msg("frontend before mux")
	mux := tracing.NewServeMux(s.Tracer)
	mux.Handle("/", http.FileServer(http.FS(staticContent)))
	mux.Handle("/hotels", http.HandlerFunc(s.searchHandler))
	mux.Handle("/recommendations", http.HandlerFunc(s.recommendHandler))
	mux.Handle("/user", http.HandlerFunc(s.userHandler))
	mux.Handle("/review", http.HandlerFunc(s.reviewHandler))
	mux.Handle("/restaurants", http.HandlerFunc(s.restaurantHandler))
	mux.Handle("/museums", http.HandlerFunc(s.museumHandler))
	mux.Handle("/cinema", http.HandlerFunc(s.cinemaHandler))
	mux.Handle("/reservation", http.HandlerFunc(s.reservationHandler))

	log.Trace().Msg("frontend starts serving")

	tlsconfig := tls.GetHttpsOpt()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}
	if tlsconfig != nil {
		log.Info().Msg("Serving https")
		srv.TLSConfig = tlsconfig
		return srv.ListenAndServeTLS("x509/server_cert.pem", "x509/server_key.pem")
	} else {
		log.Info().Msg("Serving http")
		return srv.ListenAndServe()
	}
}

func (s *Server) initSearchClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.searchClient = search.NewSearchClient(conn)
	return nil
}

func (s *Server) initReviewClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.reviewClient = review.NewReviewClient(conn)
	return nil
}

func (s *Server) initAttractionsClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.attractionsClient = attractions.NewAttractionsClient(conn)
	return nil
}

func (s *Server) initProfileClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.profileClient = profile.NewProfileClient(conn)
	return nil
}

func (s *Server) initRecommendationClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.recommendationClient = recommendation.NewRecommendationClient(conn)
	return nil
}

func (s *Server) initUserClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.userClient = user.NewUserClient(conn)
	return nil
}

func (s *Server) initReservation(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.reservationClient = reservation.NewReservationClient(conn)
	return nil
}

func (s *Server) getGprcConn(name string) (*grpc.ClientConn, error) {
	log.Info().Msg("get Grpc conn is :")
	log.Info().Msg(s.KnativeDns)
	log.Info().Msg(fmt.Sprintf("%s.%s", name, s.KnativeDns))

	if s.KnativeDns != "" {
		return dialer.Dial(
			fmt.Sprintf("consul://%s/%s.%s", s.ConsulAddr, name, s.KnativeDns),
			dialer.WithTracer(s.Tracer))
	} else {
		return dialer.Dial(
			fmt.Sprintf("consul://%s/%s", s.ConsulAddr, name),
			dialer.WithTracer(s.Tracer),
			dialer.WithBalancer(s.Registry.Client),
		)
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	log.Trace().Msg("starts searchHandler")

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// lan/lon from query params
	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}

	Lat, _ := strconv.ParseFloat(sLat, 32)
	lat := float32(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 32)
	lon := float32(Lon)

	log.Trace().Msg("starts searchHandler querying downstream")

	log.Trace().Msgf("SEARCH [lat: %v, lon: %v, inDate: %v, outDate: %v", lat, lon, inDate, outDate)
	// search for best hotels
	searchResp, err := s.searchClient.Nearby(ctx, &search.NearbyRequest{
		Lat:     lat,
		Lon:     lon,
		InDate:  inDate,
		OutDate: outDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Trace().Msg("SearchHandler gets searchResp")
	//for _, hid := range searchResp.HotelIds {
	//	log.Trace().Msgf("Search Handler hotelId = %s", hid)
	//}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	reservationResp, err := s.reservationClient.CheckAvailability(ctx, &reservation.Request{
		CustomerName: "",
		HotelId:      searchResp.HotelIds,
		InDate:       inDate,
		OutDate:      outDate,
		RoomNumber:   1,
	})
	if err != nil {
		log.Error().Msg("SearchHandler CheckAvailability failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Trace().Msgf("searchHandler gets reserveResp")
	log.Trace().Msgf("searchHandler gets reserveResp.HotelId = %s", reservationResp.HotelId)

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: reservationResp.HotelId,
		Locale:   locale,
	})
	if err != nil {
		log.Error().Msg("SearchHandler GetProfiles failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Trace().Msg("searchHandler gets profileResp")

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}

func (s *Server) recommendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}
	Lat, _ := strconv.ParseFloat(sLat, 64)
	lat := float64(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 64)
	lon := float64(Lon)

	require := r.URL.Query().Get("require")
	if require != "dis" && require != "rate" && require != "price" {
		http.Error(w, "Please specify require params", http.StatusBadRequest)
		return
	}

	// recommend hotels
	recResp, err := s.recommendationClient.GetRecommendations(ctx, &recommendation.Request{
		Require: require,
		Lat:     float64(lat),
		Lon:     float64(lon),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: recResp.HotelIds,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}

func (s *Server) reviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Logged-in successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	revInput := review.Request{HotelId: hotelId}

	revResp, err := s.reviewClient.GetReviews(ctx, &revInput)

	str = "Have reviews = " + strconv.Itoa(len(revResp.Reviews))
	if len(revResp.Reviews) == 0 {
		str = "Failed. No Reviews. "
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) restaurantHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Logged-in successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	revInput := attractions.Request{HotelId: hotelId}

	revResp, err := s.attractionsClient.NearbyRest(ctx, &revInput)

	str = "Have restaurants = " + strconv.Itoa(len(revResp.AttractionIds))
	if len(revResp.AttractionIds) == 0 {
		str = "Failed. No Restaurants. "
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) museumHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Logged-in successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	revInput := attractions.Request{HotelId: hotelId}

	revResp, err := s.attractionsClient.NearbyMus(ctx, &revInput)

	str = "Have museums = " + strconv.Itoa(len(revResp.AttractionIds))
	if len(revResp.AttractionIds) == 0 {
		str = "Failed. No Museums. "
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) cinemaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Logged-in successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	revInput := attractions.Request{HotelId: hotelId}

	revResp, err := s.attractionsClient.NearbyCinema(ctx, &revInput)

	str = "Have cinemas = " + strconv.Itoa(len(revResp.AttractionIds))
	if len(revResp.AttractionIds) == 0 {
		str = "Failed. No Cinemas. "
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Login successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) reservationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		http.Error(w, "Please check inDate/outDate format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	customerName := r.URL.Query().Get("customerName")
	if customerName == "" {
		http.Error(w, "Please specify customerName params", http.StatusBadRequest)
		return
	}

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	numberOfRoom := 0
	num := r.URL.Query().Get("number")
	if num != "" {
		numberOfRoom, _ = strconv.Atoi(num)
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Reserve successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	// Make reservation
	resResp, err := s.reservationClient.MakeReservation(ctx, &reservation.Request{
		CustomerName: customerName,
		HotelId:      []string{hotelId},
		InDate:       inDate,
		OutDate:      outDate,
		RoomNumber:   int32(numberOfRoom),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resResp.HotelId) == 0 {
		str = "Failed. Already reserved. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

// return a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSONResponse(hs []*profile.Hotel) map[string]interface{} {
	fs := []interface{}{}

	for _, h := range hs {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float32{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}

func checkDataFormat(date string) bool {
	if len(date) != 10 {
		return false
	}
	for i := 0; i < 10; i++ {
		if i == 4 || i == 7 {
			if date[i] != '-' {
				return false
			}
		} else {
			if date[i] < '0' || date[i] > '9' {
				return false
			}
		}
	}
	return true
}

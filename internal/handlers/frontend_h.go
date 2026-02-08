package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bookstore/internal/logic"
	"bookstore/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type FrontendHandler struct {
	tpls      map[string]*template.Template
	books     *logic.BookService
	auth      *logic.AuthService
	cart      *logic.CartCRUDService
	orderSvc  *logic.OrderService
	orderCRUD *logic.OrderCRUDService
	wishlist  *logic.WishlistService

	secret []byte
}

func parsePage(base string, page string) (*template.Template, error) {
	return template.ParseFiles(
		"web/templates/"+base,
		"web/templates/"+page,
	)
}

func NewFrontendHandler(
	books *logic.BookService,
	auth *logic.AuthService,
	cart *logic.CartCRUDService,
	orderSvc *logic.OrderService,
	orderCRUD *logic.OrderCRUDService,
	wishlist *logic.WishlistService,
	secret string,
) (*FrontendHandler, error) {
	if secret == "" {
		return nil, errors.New("JWT secret empty")
	}

	pages := map[string]string{
		"home":          "home.html",
		"catalog":       "catalog.html",
		"about":         "about.html",
		"login":         "login.html",
		"register":      "register.html",
		"cart":          "cart.html",
		"orders":        "orders.html",
		"order_details": "order_details.html",
		"wishlists":     "wishlists.html",
		"create_book":   "create_book.html",
	}

	tpls := make(map[string]*template.Template, len(pages))
	for key, file := range pages {
		t, err := parsePage("base.html", file)
		if err != nil {
			return nil, fmt.Errorf("parse templates for %s (%s): %w", key, file, err)
		}
		tpls[key] = t
	}

	return &FrontendHandler{
		tpls:      tpls,
		books:     books,
		auth:      auth,
		cart:      cart,
		orderSvc:  orderSvc,
		orderCRUD: orderCRUD,
		wishlist:  wishlist,
		secret:    []byte(secret),
	}, nil
}

func (h *FrontendHandler) render(w http.ResponseWriter, pageKey string, data any) {
	t, ok := h.tpls[pageKey]
	if !ok {
		http.Error(w, "template not found: "+pageKey, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func greetingByHour() string {
	hh := time.Now().Hour()
	switch {
	case hh >= 5 && hh < 12:
		return "☀️ Good morning!"
	case hh >= 12 && hh < 18:
		return "🌤️ Good afternoon!"
	case hh >= 18 && hh < 23:
		return "🌙 Good evening!"
	default:
		return "🌙 Hello!"
	}
}

func (h *FrontendHandler) setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func (h *FrontendHandler) clearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func (h *FrontendHandler) currentUser(r *http.Request) (userID int, role string, ok bool) {
	c, err := r.Cookie("token")
	if err != nil || c.Value == "" {
		return 0, "", false
	}

	tok, err := jwt.Parse(c.Value, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.secret, nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return 0, "", false
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", false
	}

	idf, ok := claims["userId"].(float64)
	if !ok {
		return 0, "", false
	}

	roleStr, _ := claims["role"].(string)
	return int(idf), roleStr, true
}

func (h *FrontendHandler) baseData(r *http.Request, active string) map[string]any {
	_, role, ok := h.currentUser(r)
	return map[string]any{
		"Greeting": greetingByHour(),
		"IsAuth":   ok,
		"Role":     role,
		"Active":   active,
	}
}

func (h *FrontendHandler) requireAuth(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, _, ok := h.currentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return 0, false
	}
	return userID, true
}

func (h *FrontendHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, role, ok := h.currentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return 0, false
	}
	if role != "admin" {
		http.Error(w, "forbidden (admin only)", http.StatusForbidden)
		return 0, false
	}
	return userID, true
}

func (h *FrontendHandler) ensureUserCart(userID int) (models.Cart, []models.CartItem) {
	all := h.cart.ListCarts()
	var found models.Cart
	for _, c := range all {
		if c.CustomerID == userID {
			found = c
			break
		}
	}
	if found.ID == 0 {
		found = h.cart.CreateCart(userID)
	}
	c, items, err := h.cart.GetCart(found.ID)
	if err != nil {
		return found, []models.CartItem{}
	}
	return c, items
}

func (h *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "home")
	data["Title"] = "Home"
	h.render(w, "home", data)
}

func (h *FrontendHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "catalog")
	data["Title"] = "Catalog"

	qp := r.URL.Query()

	var q models.BookQuery
	q.Search = strings.TrimSpace(qp.Get("search"))
	q.Genre = strings.TrimSpace(qp.Get("genre"))
	q.SortBy = strings.TrimSpace(qp.Get("sort"))
	q.Order = strings.TrimSpace(qp.Get("order"))

	if v := strings.TrimSpace(qp.Get("minPrice")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			q.MinPrice = &f
		}
	}
	if v := strings.TrimSpace(qp.Get("maxPrice")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			q.MaxPrice = &f
		}
	}

	books, err := h.books.ListBooks(r.Context(), q)
	if err != nil {
		data["Books"] = []models.Book{}
		data["Error"] = "Failed to load books"
	} else {
		data["Books"] = books
	}

	data["Q"] = map[string]string{
		"search":   q.Search,
		"genre":    q.Genre,
		"minPrice": qp.Get("minPrice"),
		"maxPrice": qp.Get("maxPrice"),
		"sort":     q.SortBy,
		"order":    q.Order,
	}

	data["Genres"] = []string{
		"",
		"Fantasy",
		"Romance",
		"Science Fiction",
		"Mystery",
		"Thriller",
		"Non-Fiction",
		"Self-Help",
		"History",
		"Biography",
		"Programming",
	}

	h.render(w, "catalog", data)
}

func (h *FrontendHandler) About(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "about")
	data["Title"] = "About"
	h.render(w, "about", data)
}

func (h *FrontendHandler) Login(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "login")
	data["Title"] = "Login"
	h.render(w, "login", data)
}

func (h *FrontendHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	pass := r.FormValue("password")

	token, err := h.auth.Login(email, pass)
	if err != nil {
		data := h.baseData(r, "login")
		data["Title"] = "Login"
		data["Error"] = "Invalid email or password"
		h.render(w, "login", data)
		return
	}

	h.setTokenCookie(w, token)
	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func (h *FrontendHandler) Register(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "register")
	data["Title"] = "Register"
	h.render(w, "register", data)
}

func (h *FrontendHandler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	pass := r.FormValue("password")

	if err := h.auth.Register(email, pass); err != nil {
		data := h.baseData(r, "register")
		data["Title"] = "Register"
		data["Error"] = err.Error()
		h.render(w, "register", data)
		return
	}

	token, err := h.auth.Login(email, pass)
	if err == nil {
		h.setTokenCookie(w, token)
	}
	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func (h *FrontendHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearTokenCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *FrontendHandler) CartPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	c, items := h.ensureUserCart(userID)

	books, _ := h.books.ListBooks(r.Context(), models.BookQuery{})
	bookMap := map[int]models.Book{}
	for _, b := range books {
		bookMap[b.ID] = b
	}

	type row struct {
		Item    models.CartItem
		Book    models.Book
		Line    float64
		HasBook bool
	}

	rows := make([]row, 0, len(items))
	var total float64

	for _, it := range items {
		b, exists := bookMap[it.BookID]

		line := 0.0
		if exists {
			line = b.Price * float64(it.Qty)
			total += line
		}

		rows = append(rows, row{
			Item:    it,
			Book:    b,
			Line:    line,
			HasBook: exists,
		})
	}

	data := h.baseData(r, "cart")
	data["Title"] = "Cart"
	data["Cart"] = c
	data["Rows"] = rows
	data["Total"] = total

	h.render(w, "cart", data)
}

func (h *FrontendHandler) CartAdd(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	bookID, _ := strconv.Atoi(r.PathValue("bookId"))
	if bookID <= 0 {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}

	c, _ := h.ensureUserCart(userID)
	_, _ = h.cart.AddItem(c.ID, bookID, 1)
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *FrontendHandler) CartUpdateQty(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	itemID, _ := strconv.Atoi(r.PathValue("itemId"))
	if itemID <= 0 {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	if qty <= 0 {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	c, _ := h.ensureUserCart(userID)
	_ = h.cart.UpdateItem(c.ID, itemID, qty)
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *FrontendHandler) CartDeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	itemID, _ := strconv.Atoi(r.PathValue("itemId"))
	if itemID <= 0 {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	c, _ := h.ensureUserCart(userID)
	_ = h.cart.DeleteItem(c.ID, itemID)
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *FrontendHandler) OrdersPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	all := h.orderCRUD.ListOrders()
	out := make([]models.Order, 0)
	for _, o := range all {
		if o.CustomerID == userID {
			out = append(out, o)
		}
	}

	data := h.baseData(r, "orders")
	data["Title"] = "Orders"
	data["Orders"] = out
	h.render(w, "orders", data)
}

func (h *FrontendHandler) OrderDetailsPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	id, _ := strconv.Atoi(r.PathValue("id"))
	if id <= 0 {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	o, items, err := h.orderCRUD.GetOrder(id)
	if err != nil || o.CustomerID != userID {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	books, _ := h.books.ListBooks(r.Context(), models.BookQuery{})

	bookMap := map[int]models.Book{}
	for _, b := range books {
		bookMap[b.ID] = b
	}

	type row struct {
		Item models.OrderItem
		Book models.Book
	}

	rows := make([]row, 0, len(items))
	for _, it := range items {
		rows = append(rows, row{Item: it, Book: bookMap[it.BookID]})
	}

	data := h.baseData(r, "orders")
	data["Title"] = "Order Details"
	data["Order"] = o
	data["Rows"] = rows
	h.render(w, "order_details", data)
}

func (h *FrontendHandler) CreateOrderFromCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	c, items := h.ensureUserCart(userID)
	if len(items) == 0 {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	_, _, _ = h.orderSvc.CreateOrderFromCart(userID, c.ID)
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

func (h *FrontendHandler) WishlistsPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	all := h.wishlist.ListWishlists()

	var myWL models.Wishlist
	for _, wli := range all {
		if wli.CustomerID == userID {
			myWL = wli
			break
		}
	}
	if myWL.ID == 0 {
		myWL = h.wishlist.CreateWishlist(userID)
		all = h.wishlist.ListWishlists()
	}

	books, _ := h.books.ListBooks(r.Context(), models.BookQuery{})
	bookMap := map[int]models.Book{}
	for _, b := range books {
		bookMap[b.ID] = b
	}

	myObj, myItems, _ := h.wishlist.GetWishlist(myWL.ID)
	myRows := make([]WishlistRowView, 0, len(myItems))
	var myTotal float64
	for _, it := range myItems {
		b := bookMap[it.BookID]
		line := b.Price * float64(it.Qty)
		myTotal += line
		myRows = append(myRows, WishlistRowView{
			Item: it,
			Book: b,
			Line: line,
		})
	}
	myBlock := WishlistBlockView{
		Wishlist: myObj,
		Rows:     myRows,
		Total:    myTotal,
	}

	others := make([]WishlistBlockView, 0)
	for _, wl := range all {
		if wl.CustomerID == userID {
			continue
		}

		wObj, items, err := h.wishlist.GetWishlist(wl.ID)
		if err != nil {
			continue
		}

		rows := make([]WishlistRowView, 0, len(items))
		var total float64
		for _, it := range items {
			b := bookMap[it.BookID]
			line := b.Price * float64(it.Qty)
			total += line
			rows = append(rows, WishlistRowView{
				Item: it,
				Book: b,
				Line: line,
			})
		}

		others = append(others, WishlistBlockView{
			Wishlist: wObj,
			Rows:     rows,
			Total:    total,
		})
	}

	data := h.baseData(r, "wishlists")
	data["Title"] = "Wishlists"
	data["My"] = myBlock
	data["Others"] = others

	h.render(w, "wishlists", data)
}

func (h *FrontendHandler) WishlistAdd(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	bookID, _ := strconv.Atoi(r.PathValue("bookId"))
	if bookID <= 0 {
		http.Redirect(w, r, "/wishlists", http.StatusSeeOther)
		return
	}

	all := h.wishlist.ListWishlists()
	var wl models.Wishlist
	for _, wli := range all {
		if wli.CustomerID == userID {
			wl = wli
			break
		}
	}
	if wl.ID == 0 {
		wl = h.wishlist.CreateWishlist(userID)
	}

	_, _ = h.wishlist.AddItem(wl.ID, bookID, 1)
	http.Redirect(w, r, "/wishlists", http.StatusSeeOther)
}

func (h *FrontendHandler) WishlistGift(w http.ResponseWriter, r *http.Request) {
	buyerID, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	wishlistID, _ := strconv.Atoi(r.PathValue("wishlistId"))
	if wishlistID <= 0 {
		http.Redirect(w, r, "/wishlists", http.StatusSeeOther)
		return
	}

	_, _, _, _ = h.wishlist.GiftFromWishlist(wishlistID, buyerID)
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}
func (h *FrontendHandler) AdminCreateBookPage(w http.ResponseWriter, r *http.Request) {
	_, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	data := h.baseData(r, "admin")
	data["Title"] = "Admin: Create Book"

	data["Form"] = map[string]string{
		"title":       "",
		"author":      "",
		"genre":       "",
		"price":       "0.00",
		"description": "",
	}

	h.render(w, "create_book", data)
}

func (h *FrontendHandler) AdminCreateBookPost(w http.ResponseWriter, r *http.Request) {
	_, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	_ = r.ParseForm()

	title := strings.TrimSpace(r.FormValue("title"))
	author := strings.TrimSpace(r.FormValue("author"))
	genre := strings.TrimSpace(r.FormValue("genre"))
	priceStr := strings.TrimSpace(r.FormValue("price"))
	desc := strings.TrimSpace(r.FormValue("description"))

	form := map[string]string{
		"title":       title,
		"author":      author,
		"genre":       genre,
		"price":       priceStr,
		"description": desc,
	}

	if title == "" || author == "" || genre == "" || priceStr == "" || desc == "" {
		data := h.baseData(r, "admin")
		data["Title"] = "Admin: Create Book"
		data["Error"] = "All fields are required."
		data["Form"] = form
		h.render(w, "create_book", data)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		data := h.baseData(r, "admin")
		data["Title"] = "Admin: Create Book"
		data["Error"] = "Price must be a valid number."
		data["Form"] = form
		h.render(w, "create_book", data)
		return
	}
	if price <= 0 {
		data := h.baseData(r, "admin")
		data["Title"] = "Admin: Create Book"
		data["Error"] = "Price must be greater than 0."
		data["Form"] = form
		h.render(w, "create_book", data)
		return
	}

	_, err = h.books.CreateBook(models.Book{
		Title:       title,
		Author:      author,
		Genre:       genre,
		Price:       price,
		Description: desc,
	})
	if err != nil {
		data := h.baseData(r, "admin")
		data["Title"] = "Admin: Create Book"
		data["Error"] = err.Error()
		data["Form"] = form
		h.render(w, "create_book", data)
		return
	}

	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

type WishlistRowView struct {
	Item models.WishlistItem
	Book models.Book
	Line float64
}

type WishlistBlockView struct {
	Wishlist models.Wishlist
	Rows     []WishlistRowView
	Total    float64
}

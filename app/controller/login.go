package controller

import (
    "fmt"
    "log"
    "net/http"

    "github.com/xusheng6/crackmes.one/app/model"
    "github.com/xusheng6/crackmes.one/app/shared/passhash"
    "github.com/xusheng6/crackmes.one/app/shared/session"
    "github.com/xusheng6/crackmes.one/app/shared/view"

    "github.com/gorilla/sessions"
    "github.com/josephspurrier/csrfbanana"
)

const (
    // Name of the session variable that tracks login attempts
    sessLoginAttempt = "login_attempt"
)

// loginAttempt increments the number of login attempts in sessions variable
func loginAttempt(sess *sessions.Session) {
    // Log the attempt
    if sess.Values[sessLoginAttempt] == nil {
        sess.Values[sessLoginAttempt] = 1
    } else {
        sess.Values[sessLoginAttempt] = sess.Values[sessLoginAttempt].(int) + 1
    }
}

// LoginGET displays the login page
func LoginGET(w http.ResponseWriter, r *http.Request) {
    // Get session
    sess := session.Instance(r)

    // Display the view
    v := view.New(r)
    v.Name = "login/login"
    v.Vars["token"] = csrfbanana.Token(w, r, sess)
    // Refill any form fields
    view.Repopulate([]string{"email"}, r.Form, v.Vars)
    v.Render(w)
}

// LoginPOST handles the login form submission
func LoginPOST(w http.ResponseWriter, r *http.Request) {
    // Get session
    sess := session.Instance(r)

    // Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
    /*if sess.Values[sessLoginAttempt] != nil && sess.Values[sessLoginAttempt].(int) >= 5 {
        log.Println("Brute force login prevented")
        sess.AddFlash(view.Flash{"Sorry, no brute force :-)", view.FlashNotice})
        sess.Save(r, w)
        LoginGET(w, r)
        return
    }*/

    // Validate with required fields
    if validate, missingField := view.Validate(r, []string{"name", "password"}); !validate {
        sess.AddFlash("Field missing: " + missingField)
        sess.Save(r, w)
        LoginGET(w, r)
        return
    }

    // Form values
    name := r.FormValue("name")
    password := r.FormValue("password")

    if !view.AuthorizedCharsOnly(name){
        sess.AddFlash(view.Flash{"Non authorized chars", view.FlashError})
        sess.Save(r, w)
        LoginGET(w, r)
        return
    }

    // Get database result
    result, err := model.UserByName(name)

    // Determine if user exists
    if err == model.ErrNoResult {
        loginAttempt(sess)
        sess.AddFlash(view.Flash{"Password is incorrect - Attempt: " + fmt.Sprintf("%v", sess.Values[sessLoginAttempt]), view.FlashWarning})
        sess.Save(r, w)
    } else if err != nil {
        // Display error message
        log.Println(err)
        sess.AddFlash(view.Flash{"There was an error. Please try again later.", view.FlashError})
        sess.Save(r, w)
    } else if passhash.MatchString(result.Password, password) {
        // Login successfully
        session.Empty(sess)
        sess.AddFlash(view.Flash{"Login successful!", view.FlashSuccess})
        sess.Values["email"] = result.Email
        sess.Values["name"] = result.Name
        sess.Save(r, w)
        http.Redirect(w, r, "/", http.StatusFound)
        return
    } else {
        loginAttempt(sess)
        sess.AddFlash(view.Flash{"Password is incorrect - Attempt: " + fmt.Sprintf("%v", sess.Values[sessLoginAttempt]), view.FlashWarning})
        sess.Save(r, w)
    }

    // Show the login page again
    LoginGET(w, r)
}

// LogoutGET clears the session and logs the user out
func LogoutGET(w http.ResponseWriter, r *http.Request) {
    // Get session
    sess := session.Instance(r)

    // If user is authenticated
    if sess.Values["name"] != nil {
        session.Empty(sess)
        sess.AddFlash(view.Flash{"Goodbye!", view.FlashNotice})
        sess.Save(r, w)
    }

    http.Redirect(w, r, "/", http.StatusFound)
}

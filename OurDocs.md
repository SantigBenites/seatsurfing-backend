# Goal

Make weekly reservations of the same room for a amount of time (probably 6 months)

## Found so far on frontend

In class search.tsx there is a method
```
onConfirmBooking = () => {
if (this.state.selectedSpace == null) {
    return;
}
this.setState({
    showConfirm: false,
    loading: true
});
let booking: Booking = new Booking();
booking.enter = new Date(this.state.enter);
booking.leave = new Date(this.state.leave);
if (!RuntimeConfig.INFOS.dailyBasisBooking) {
    booking.leave.setSeconds(booking.leave.getSeconds() - 1);
}
booking.space = this.state.selectedSpace;
booking.save().then(() => {
    this.setState({
    loading: false,
    showSuccess: true
    });
}).catch(e => {
    let code: number = 0;
    if (e instanceof AjaxError) {
    code = e.appErrorCode;
    }
    this.setState({
    loading: false,
    showError: true,
    errorText: ErrorText.getTextForAppCode(code, this.props.t)
    });
});
}
```

Which saves the booking made by the user

And this method save is defined as

```
async save(): Promise<Booking> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
}
```

## Found so far on backend

backend url for bookings:
```
getBackendUrl(): string {
    return "/booking/";
}
```

Save entity method of Ajax

```
static async saveEntity(e: Entity, url: string): Promise<AjaxResult> {
    if (!url.endsWith("/")) {
        url += "/";
    }
    if (e.id) {
        return Ajax.putData(url + e.id, e.serialize());
    } else {
        return Ajax.postData(url, e.serialize()).then(result => {
        e.id = result.objectId;
        return result;
        });
    }
}
```

Put the data in the database

```
static async putData(url: string, data?: any): Promise<AjaxResult> {
    return Ajax.query("PUT", url, data);
}
```

The actual query function

```

static async query(method: string, url: string, data?: any): Promise<AjaxResult> {
    console.log("AJAX request for: " + url);
    url = Ajax.getBackendUrl() + url;
    return new Promise<AjaxResult>(function (resolve, reject) {
        let performRequest = () => {
        let options: RequestInit = Ajax.getFetchOptions(method, Ajax.CREDENTIALS.accessToken, data);
        fetch(url, options).then((response) => {
            if (response.status >= 200 && response.status <= 299) {
                response.json().then(json => {
                    resolve(Ajax.getAjaxResult(json, response));
                }).catch(err => {
                    resolve(Ajax.getAjaxResult({}, response));
                });
            } else {
                let appCode = response.headers.get("X-Error-Code");
                reject(new AjaxError(response.status, appCode ? parseInt(appCode) : 0));
            }
        }).catch(err => {
            reject(err);
        });
        };
        if (Ajax.CREDENTIALS.refreshToken) {
            Ajax.refreshAccessToken(Ajax.CREDENTIALS.refreshToken).then(cred => {
                performRequest();
            }).catch(err => {
                reject(new AjaxError(401, 0));
            });
        } else {
            performRequest();
        }
    });
}
```

## Found so far on actual server

Ajax then calls the server with POST method of /booking/

```
func (router *BookingRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	space, err := GetSpaceRepository().GetOne(m.SpaceID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	e, err := router.copyFromRestModel(&m, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	e.UserID = GetRequestUserID(r)
	if m.UserEmail != "" && m.UserEmail != requestUser.Email {
		if !CanSpaceAdminOrg(requestUser, location.OrganizationID) {
			SendForbidden(w)
			return
		}
		e.UserID, err = router.bookForUser(requestUser, m.UserEmail, w)
		if err != nil {
			SendInternalServerError(w)
			return
		}
	}
	bookingReq := &BookingRequest{
		Enter: e.Enter,
		Leave: e.Leave,
	}

	if valid, code := router.checkBookingCreateUpdate(bookingReq, location, requestUser, ""); !valid {
		log.Println(err)
		SendBadRequestCode(w, code)
		return
	}
	conflicts, err := GetBookingRepository().GetConflicts(e.SpaceID, e.Enter, e.Leave, "")
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		SendAleadyExists(w)
		return
	}
	if err := GetBookingRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}
```

# Solution

Step by step

First step:
 - <del> Make a button that allows us to choose between weekly reservations or not </del>

Only doing in case button is true

Second step
 - <del> Appear a new date box similar to the one that already exist in main page (http://di-seatsurfing.di.fc.ul.pt:8080/ui/search) <img src="data_image.png" alt="drawing" width="250"/> </del>

Third step
 - Calculate all the dates that the room will be booked and store them in an array

Fourth step
 - Change onConfirmBooking to make the exactly same but for each of the dates on the array

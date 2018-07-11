function login()
{
	var apiUrl = "https://nuts.rtradetechnologies.com:6767/api/v1/login";
	var address = document.getElementById("ethAddress").value;
	var password = document.getElementById("ethPassword").value;
	console.log(address);
	console.log(password);
	
	//send post request to API
	var request = new XMLHttpRequest();
	request.open('POST', apiUrl, true);
	request.setRequestHeader("Cache-Control", "no-cache");
	request.setRequestHeader("Content-Type", "application/json");
	
	var formData = JSON.stringify({
        "username": address,
        "password": password
		});
	
	request.onload = function ()
    {
        if(request.status < 400)
        {
			//login was successful
			var data = JSON.parse(this.response);
			console.log(data);
			window.sessionStorage.token = data.token;
			$("#modalSuccess").modal()
        }
        else
        {
            console.log("Error logging in");
			console.log(this.response);
			$("#modalFail").modal()
        }
    }
	request.onerror = function ()
	{
		console.log(request.responseText);
    }
    request.send(formData);
}
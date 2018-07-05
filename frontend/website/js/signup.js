function register()
{
	var apiUrl = "https://nuts.rtradetechnologies.com:6767/api/v1/register";
	// this is how we can get all elements  attached to id "registration"
	var address = document.getElementById("ethAddress").value;
	var password = document.getElementById("ethPassword").value;
	var email = document.getElementById("email").value;
	console.log(address);
	console.log(password);
	console.log(email);
	
	//send post request to API
	var request = new XMLHttpRequest();
	request.open('POST', apiUrl, true);
	request.setRequestHeader("Cache-Control", "no-cache");
	
	var formData = new FormData();
	formData.append("eth_address", address);
	formData.append("password", password);
	formData.append("email_address", email);
	
	request.onload = function ()
    {
        if(request.status < 400)
        {
			//register was successful
			var data = JSON.parse(this.response);
			console.log(data);
			$("#modalSuccess").modal()
        }
        else
        {
            console.log("Error registering");
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
function pin()
{
	if(window.sessionStorage.token == undefined || window.sessionStorage.token == "")
	{
		window.alert("Not logged in");
	}
	else
	{
		//variables and get user input
		var hash = document.getElementById("contentHash").value;
		var holdTime = document.getElementById("holdTime").value;
		var apiUrl = "https://nuts.rtradetechnologies.com:6767/api/v1/ipfs/pin/" + hash;
		
		//send api request
		var request = new XMLHttpRequest();
		request.open('POST', apiUrl, true);
		request.setRequestHeader("Cache-Control", "no-cache");
		request.setRequestHeader('Authorization', 'Bearer ' + window.sessionStorage.token );
		
		var formData = new FormData();
		formData.append("hold_time", holdTime);
		
		request.onload = function ()
		{
			if(request.status < 400)
			{
				//pin was successful
				var data = JSON.parse(this.response);
				console.log(data);
				window.alert("Pin Successful");
			}
			else
			{
				console.log("Error pinning");
				console.log(this.response);
				window.alert("Pin failed");
			}
		}
		request.onerror = function ()
		{
			console.log(request.responseText);
		}
		request.send(formData);
	}
}
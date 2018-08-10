function pin()
{
	if(window.sessionStorage.token == undefined || window.sessionStorage.token == "")
	{
		window.alert("Not logged in");
	}
	else
	{
		//variables and get user input
		var keyType = document.getElementById("keyType").value;
        var keyBits = document.getElementById("keyBits").value;
        var keyName = document.getElementById("keyName".value);
		var apiUrl = "https://nuts.rtradetechnologies.com:6767/api/v1/account/key/ipfs/new";
		
		//send api request
		var request = new XMLHttpRequest();
		request.open('POST', apiUrl, true);
		request.setRequestHeader("Cache-Control", "no-cache");
		request.setRequestHeader('Authorization', 'Bearer ' + window.sessionStorage.token );
		
		var formData = new FormData();
		formData.append("key_type", keyType);
        formData.append("key_bits", keyBits);
        formData.append("key_name", keyName);
        
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
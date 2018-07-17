function upload()
{
	if(window.sessionStorage.token == undefined || window.sessionStorage.token == "")
	{
		window.alert("Not logged in");
	}
	else
	{
		//variables and get user input
		var apiUrl = "https://nuts.rtradetechnologies.com:6767/api/v1/ipfs/add-file";
		var file = document.getElementById("fileUpload").value;
		var holdTime = document.getElementById("holdTime").value;
		
		//send api request
		var request = new XMLHttpRequest();
		request.open('POST', apiUrl, true);
		request.setRequestHeader("Cache-Control", "no-cache");
		request.setRequestHeader('Authorization', 'Bearer ' + window.sessionStorage.token );
		
		var formData = new FormData();
		formData.append("file", file);
		formData.append("hold_time", holdTime);
		formData.append("use_private_network", "yes");
		formData.append("network_name", "private_net_1");
		
		request.onload = function ()
		{
			if(request.status < 400)
			{
				//pin was successful
				var data = JSON.parse(this.response);
				console.log(data);
				window.alert("Upload Successful");
			}
			else
			{
				console.log("Error uploading file");
				console.log(this.response);
				window.alert("Upload failed");
			}
		}
		request.onerror = function ()
		{
			console.log(request.responseText);
		}
		request.send(formData);
	}
}
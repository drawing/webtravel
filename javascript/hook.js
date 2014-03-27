document.domain = "{{.Domain}}"

if (top.location == self.location)
{
	var _gaq = _gaq || [];
	_gaq.push(['_setAccount', 'UA-37834254-1']);
	_gaq.push(['_trackPageview']);
	
	(function() {
	var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
	ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
	var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
	})();
}

function skip_access_convert_url_0001(oldurl)
{
	if (typeof(oldurl) != "string") {
		return oldurl
	}
	
	newurl = oldurl
	
	if (oldurl.indexOf("{{.FullURI}}") == 0) {
		return newurl
	}
	if (oldurl.indexOf("{{.RootURI}}") == 0) {
		return skip_access_convert_url_0001(oldurl.substring("{{.RootURI}}".length))
	}
	
	if (document.location.href.indexOf("{{.RootURI}}") == 0) {
		temp = document.location.href.substr("{{.RootURI}}".length)
		var pos = temp.indexOf("db.")
		if (pos != -1) {
			temp = temp.substring(0, pos+4)
			if (oldurl.indexOf(temp) == 0) {
				return skip_access_convert_url_0001(oldurl.substring(temp.length))
			}
		}
	}
	
	if (oldurl.indexOf("#") == 0) {
	} else if (oldurl.indexOf("about:blank") == 0) {
	} else if (oldurl.indexOf("data:") == 0) {
	} else if (oldurl.indexOf("file:") == 0) {
	} else if (oldurl.indexOf("res:") == 0) {
	} else if (oldurl.indexOf("C:") == 0) {
	} else if (oldurl.indexOf("javascript:") == 0) {
	}
	else if (oldurl.indexOf("http://") == 0) {
		newurl = oldurl.substr(7)
		var pos = newurl.indexOf("/")
		host = newurl
		path = "/"
		if (pos != -1) {
			host = newurl.substring(0, pos)
			path = newurl.substring(pos)
		}
		arr = host.split(".")
		arr.reverse()
		host = arr.join("/")
		newurl = "{{.FullURI}}" + host + "/db.a" + path
	}
	else if (oldurl.indexOf("https://") == 0) {
		newurl = oldurl.substr(8)
		var pos = newurl.indexOf("/")
		host = newurl
		path = "/"
		if (pos != -1) {
			host = newurl.substring(0, pos)
			path = newurl.substring(pos)
		}
		arr = host.split(".")
		arr.reverse()
		host = arr.join("/")
		newurl = "{{.FullURI}}" + host + "/db.b" + path
	}
	else if (oldurl.indexOf("//") == 0) {
		newurl = skip_access_convert_protocol_0001(window.location.protocol) + oldurl
		newurl = skip_access_convert_url_0001(newurl)
	}
	else if (oldurl.indexOf("/") == 0) {
		var pos = document.location.href.indexOf("db.")
		pos = pos + 4
		newurl = document.location.href.substring(0, pos)
		newurl = newurl + oldurl
	}
	else {
		var pos = document.location.href.lastIndexOf("/")
		if (pos != -1) {
			newurl = document.location.href.substring(0, pos+1) + oldurl
		}
	}
    return newurl
}
function skip_access_convert_domain_0001(old_domain)
{
	if (typeof(old_domain) != "string") {
		return old_domain
	}
	return "{{.Domain}}"
}
function skip_access_convert_path_0001(oldpath)
{
	// alert(oldpath)
	if (typeof(oldpath) != "string") {
		return oldpath
	}
	newpath = skip_access_convert_url_0001(oldpath)
	pos = newpath.indexOf("://")
	if (pos == -1) {
		return oldpath
	}
	newpath = newpath.substr(pos+3)
	pos = newpath.indexOf("/")
	if (pos == -1) {
		return oldpath
	}
	newpath = newpath.substr(pos)
	// alert(newpath)
	return newpath
}

function skip_access_convert_html_0001(html)
{
	if (typeof(html) != 'string')
	{
		return html;
	}

	// Extract a base tag
	if ((parser = /<base href(?==)=["']?([^"' >]+)['"]?(>|\/>|<\/base>)/i.exec(html)))
	{
		ginf.target.b = parser[1]; // Update base variable for future parsing
		if ( ginf.target.b.charAt(ginf.target.b.length-1) != '/' ) // Ensure trailing slash
			ginf.target.b += '/'; 
		html = html.replace(parser[0],''); // Remove from document since we don't want the unproxied URL
	}
	
	// Meta refresh
	if (parser = /content=(["'])?([0-9]+)\s*;\s*url=(['"]?)([^"'>]+)\3\1(.*?)(>|\/>)/i.exec(html))
		html = html.replace(parser[0],parser[0].replace(parser[4],skip_access_convert_url_0001(parser[4])));

	// Proxy an update to URL based attributes
	html = html.replace(/\.(action|src|location|href)\s*=\s*([^;}]+)/ig,'.$1=skip_access_convert_url_0001($2)');
	 
	// Send innerHTML updates through our parser
	html = html.replace(/\.innerHTML\s*(\+)?=\s*([^};]+)\s*/ig,'.innerHTML$1=skip_access_convert_html_0001($2)');
	 
	// Proxy iframe, ensuring the frame flag is added
	parser = /<iframe\s+([^>]*)\s*src\s*=\s*(["']?)([^"']+)\2/ig;
	while (match = parser.exec(html)) 
		html = html.replace(match[0],'<iframe ' +match[1] +' src'+'=' + match[2] + skip_access_convert_url_0001(match[3],'frame') + match[2] );

	// Proxy attributes
	parser = /\s(href|src|background|action)\s*=\s*(["']?)([^"'\s>]+)/ig;
	while (match = parser.exec(html))
	{
		html = html.replace(match[0], ' '+match[1]+'='+match[2]+skip_access_convert_url_0001(match[3]));
	}
	
	// Convert get to post
	// parser = /<fo(?=r)rm((?:(?!method)[^>])*)(?:\s*method\s*=\s*(["']?)(get|post)\2)?([^>]*)>/ig;
	// while (match = parser.exec(html))
	//{
	//	if (!match[3] || match[3].toLowerCase() != 'post')
	//		html = html.replace(match[0],'<fo'+'rm'+match[1]+' method="post" '+match[4]+'><input type="hidden" name="convertGET" value="1">');
	//}
	
	// Proxy CSS: url(someurl.com/image.gif)
	parser = /url\s*\(['"]?([^'"\)]+)['"]?\)/ig;
	while (match = parser.exec(html))
	{
		html = html.replace(match[0],'url('+skip_access_convert_url_0001(match[1])+')');
	}

	// Proxy CSS importing stylesheets
	parser = /@import\s*['"]([^'"\(\)]+)['"]/ig;
	while (match = parser.exec(html))
	{
		html = html.replace(match[0],'@import "'+skip_access_convert_url_0001(match[1])+'"');
	}

	// Return changed HTML
	return html;
}

/*
 * Ajax XMLHttpRequest Open Hook
 */
if (typeof(XMLHttpRequest) != "undefined" &&
	typeof(XMLHttpRequest.prototype) != "undefined")
{
	var oriXOpen = XMLHttpRequest.prototype.open; 
	XMLHttpRequest.prototype.open = function(method, url, asncFlag, user, password) {
		url = skip_access_convert_url_0001(url)
		oriXOpen.call(this,method, url, asncFlag, user, password); 
	};
}

if (typeof(XMLDocument) != "undefined" &&
	typeof(XMLDocument.prototype) != "undefined")
{
	var oriDLoad = XMLDocument.prototype.load
	XMLDocument.prototype.load = function(filePath)
	{
		// alert("filepath:" + filePath)
		filePath = skip_access_convert_url_0001(filePath)
		oriDLoad.call(this, filePath)
	}
}

if (typeof(HTMLDocument) != "undefined" &&
	typeof(HTMLDocument.prototype) != "undefined")
{
	var oriHWrite = HTMLDocument.prototype.write
	HTMLDocument.prototype.write = function(s)
	{
		// alert("Old document.Write: " + s)
		ns = skip_access_convert_html_0001(s)
		// alert("New document.Write: " + ns)
		oriHWrite.call(this, ns)
	}
}

function skip_access_setAttribute_0001(name, val) {
	if (typeof(name) != "string" || typeof(val) != "string") {
		return val
	}

	newValue = val;
	//if (val == "http://us-st.xhamster.com/videoplayer3.swf") {
	//	return "{{.FullURI}}" + "com/xhamster/us-st/db.a/videoplayer3.swf";
	//}
	
	if(name == "src" || name == "action" || name == "href" || name == "data") {
        newValue = skip_access_convert_url_0001(val)
    }
	return newValue
}

function skip_access_convert_host_0001(oldhost)
{
	// alert(oldhost)
	if (oldhost != window.location.host)
	{
		return oldhost;
	}
	var pos = window.location.href.indexOf("/db.")
	if (pos == -1) {
		return oldhost;
	}
	newhost = window.location.href.substring(0, pos)
	
	if (newhost.indexOf("{{.FullURI}}") != 0) {
		return oldhost
	}
	newhost = newhost.substring("{{.FullURI}}".length)
	
	arr = newhost.split("/")
	arr.reverse()
	newhost = arr.join(".")

	return newhost
}

function skip_access_convert_protocol_0001(oldprotocol)
{
	if (oldprotocol != window.location.protocol)
	{
		return oldprotocol;
	}
	var pos = document.location.href.indexOf("db.")
	if (pos == -1) {
		return oldprotocol
	}
	newprotocol = document.location.href.substr(pos + 3)
	if (newprotocol.indexOf("a") == 0) {
		newprotocol = "http:"
	} else if (newprotocol.indexOf("b") == 0) {
		newprotocol = "https:"
	} else {
		newprotocol = oldprotocol
	}
	
	return newprotocol
}
function skip_access_convert_cookie_0001(oldcookie)
{
	// alert(oldcookie)
	// arr = oldcookie.split(";")
	return oldcookie
}

function skip_access_regex_test_0001(target)
{
	// alert(target)
	if (target != document.location) {
		return target
	}
	var pos = window.location.href.indexOf("/db.")
	if (pos == -1) {
		return target;
	}
	pos += 5
	newtarget = skip_access_convert_protocol_0001(window.location.protocol) + "//"
			+ skip_access_convert_host_0001(window.location.host)
			+ window.location.href.substr(pos)
	// alert(newtarget)
	return newtarget
}


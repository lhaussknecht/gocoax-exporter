function $(id){return document.getElementById(id);}
function _w(s){return document.write(s);}
function enableRefresh(enable){$("btnRefresh").disabled = (enable==0?true:false);}
function doRefresh(){enableRefresh(0);window.location.reload(true);}
function jump(url){window.location.href = url;}


function getCookie(name) {
  let cookieValue = null;
  if (document.cookie && document.cookie !== '') {
     const cookies = document.cookie.split(';');
     for (let i = 0; i < cookies.length; i++) {
       const cookie = cookies[i].trim();
       // Does this cookie string begin with the name we want?
       if (cookie.substring(0, name.length + 1) === (name + '=')) {
          cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
          break;
       }
     }
  } 
  return cookieValue;
}


var browser = {};

// From https://github.com/douglascrockford/JSON-js/blob/master/json2.js
if (typeof JSON.parse !== "function") {
    var rx_one = /^[\],:{}\s]*$/;
    var rx_two = /\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g;
    var rx_three = /"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g;
    var rx_four = /(?:^|:|,)(?:\s*\[)+/g;
    var rx_dangerous = /[\u0000\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g;
    JSON.parse = function(text, reviver) {
        var j;

        function walk(holder, key) {
            var k;
            var v;
            var value = holder[key];
            if (value && typeof value === "object") {
                for (k in value) {
                    if (Object.prototype.hasOwnProperty.call(value, k)) {
                        v = walk(value, k);
                        if (v !== undefined) {
                            value[k] = v;
                        } else {
                            delete value[k];
                        }
                    }
                }
            }
            return reviver.call(holder, key, value);
        }
        text = String(text);
        rx_dangerous.lastIndex = 0;
        if (rx_dangerous.test(text)) {
            text = text.replace(rx_dangerous, function (a) {
                return (
                    "\\u"
                    + ("0000" + a.charCodeAt(0).toString(16)).slice(-4)
                );
            });
        }

        if (
            rx_one.test(
                text
                    .replace(rx_two, "@")
                    .replace(rx_three, "]")
                    .replace(rx_four, "")
            )
        ) {
            j = eval("(" + text + ")");
            return (typeof reviver === "function")
                ? walk({"": j}, "")
                : j;
        }
        throw new SyntaxError("JSON.parse");
    };
}
function jump(url)
{
    window.location.href = url;
}
function doReboot()
{
    document.formAct.action = "/ms/1/0xb00"
    doFormJSON(document.formAct, "reboot.html", 1, function(retVal) {}, function(retVal) {}, 0);
    document.getElementById("container").className = "hideAll";
    document.getElementById("container2").className = "";
    setTimeout("jump('index.html')", 10000);
}
function getBrowser()
{
    var b = {};
    var ua = navigator.userAgent.toLowerCase();
    var s;
    (s = ua.match(/msie ([\d.]+)/)) ? b.ie = s[1] :
    (s = ua.match(/firefox\/([\d.]+)/)) ? b.firefox = s[1] :
    (s = ua.match(/chrome\/([\d.]+)/)) ? b.chrome = s[1] :
    (s = ua.match(/opera.([\d.]+)/)) ? b.opera = s[1] :
    (s = ua.match(/version\/([\d.]+).*safari/)) ? b.safari = s[1] : 0;

    return b;
}
function isIE(ver){browser = getBrowser();return (browser.ie && (browser.ie.indexOf(ver) != -1))?true:false;}
function isOldBrowser(){return (isIE("6.0") || isIE("7.0") || isIE("8.0"));}
function createXmlHttp()
{
    var xmlHttp = null;
    if(window.XMLHttpRequest)
        xmlHttp = new XMLHttpRequest();
    else if(window.ActiveXObject)
        xmlHttp = new ActiveXObject("Microsoft.XMLHTTP");

    return xmlHttp;
}
function getFormParam(objForm)
{
    var f = objForm;
    var aParams = new Array();
    var i = 0;
    var sParam = '';
    for(i=0; i<f.elements.length; i++)
    {
        if(f.elements[i].type == "button") continue;
        if((f.elements[i].type == "radio") && (f.elements[i].checked == false)) continue;
        sParam = encodeURIComponent(f.elements[i].name);
        sParam += "=";
        sParam += encodeURIComponent(f.elements[i].value);
        aParams.push(sParam);
    }
    return aParams.join("&");
}

function getFormParamJSON(objForm, occurence)
{
    var f = objForm;
    var aParams = new Array();
    var i = 0;
    var sParam = '';        
    var numOfData = 1;

    for(i=0; i<f.elements.length; i++)
    {
        if(f.elements[i].type == "button") continue;
        if((f.elements[i].type == "radio") && (f.elements[i].checked == false)) continue;
        sParam = "{\"";
        if (occurence == numOfData)
        {
            sParam += encodeURIComponent("data");
            sParam += "\":";
            sParam += "[";
            sParam += decodeURIComponent(f.elements[i].value);    // Used to be encode, changed to decodeURIComponent to replace %20 to spaces and %2c to ,
            sParam += "]}";
            aParams.push(sParam);
        }
        numOfData++;
    }
    return aParams.join("");
}

function getFormParamMultipleDataJSON(objForm)
{
    var f = objForm;
    var aParams = new Array();
    var i = 0;
    var sParam = '';        
    var numOfData = 0;

    sParam = "{\"";
    sParam += encodeURIComponent("data");
    sParam += "\":";
    sParam += "[";

    for(i=0; i<f.elements.length; i++)
    {
        if(f.elements[i].type == "button") continue;
        if((f.elements[i].type == "radio") && (f.elements[i].checked == false)) continue;
        if (numOfData > 0)
        {
            sParam += ",";
        }
        sParam += decodeURIComponent(f.elements[i].value);    // Used to be encode, changed to decodeURIComponent to replace %20 to spaces and %2c to ,
        numOfData++;
    }
    sParam += "]}";
    aParams.push(sParam);

    return aParams.join("");
}


function doFormJSON(objForm, url, occurence, success, error, nextIdx)
{
    var f = objForm;
    var sParam = '';
    var xmlHttp = createXmlHttp();
   
    if(xmlHttp != null)
    {
        xmlHttp.onreadystatechange = function(){
            if(xmlHttp.readyState == 4)
            {
                if(xmlHttp.status == 200)
                {
                    success(nextIdx);
                }
                else if (error)
                {
                    error(xmlHttp);
                }
            }
        };
        xmlHttp.open(f.method, f.action, true);
        xmlHttp.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
        //xmlHttp.setRequestHeader("User-Agent","MXL Browser 1.0");
        xmlHttp.setRequestHeader("Accept","text/html, */*");
        xmlHttp.setRequestHeader("X-CSRF-TOKEN", getCookie("csrf_token"));
        sParam = getFormParamJSON(f, occurence);
        xmlHttp.send(sParam);
    }
    else
    {
        alert("Fail create xmlHttp");
    }
}


function doFormGetJSON(objForm, url, success, error, curIdx)
{
    var f = objForm;
    var sParam = '';
    var xmlHttp = createXmlHttp();
    if(xmlHttp != null)
    {
        xmlHttp.onreadystatechange = function()
        {
            if(xmlHttp.readyState == 4)
            {
                if(xmlHttp.status == 200)
                {
                    if (success)
                    {
                        success(JSON.parse(xmlHttp.responseText), curIdx);
                    }
                }
                else if (error)
                {
                    error(xmlHttp);
                }
            }
        };
        xmlHttp.open(f.method, f.action, true);
        xmlHttp.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
        //xmlHttp.setRequestHeader("User-Agent","MXL Browser 1.0");
        xmlHttp.setRequestHeader("Accept","text/html, */*");
        xmlHttp.setRequestHeader("X-CSRF-TOKEN", getCookie("csrf_token"));
        sParam = getFormParamJSON(f,1);
        xmlHttp.send(sParam);
    }
    else
    {
        alert("Fail create xmlHttp");
    }
}

function doFormGetMultipleDataJSON(objForm, url, success, error, curIdx)
{
    var f = objForm;
    var sParam = '';
    var xmlHttp = createXmlHttp();
    if(xmlHttp != null)
    {
        xmlHttp.onreadystatechange = function()
        {
            if(xmlHttp.readyState == 4)
            {
                if(xmlHttp.status == 200)
                {
                    if (success)
                    {
                        success(JSON.parse(xmlHttp.responseText), curIdx);
                    }
                }
                else if (error)
                {
                    error(xmlHttp);
                }
            }
        };
        xmlHttp.open(f.method, f.action, true);
        xmlHttp.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
        //xmlHttp.setRequestHeader("User-Agent","MXL Browser 1.0");
        xmlHttp.setRequestHeader("Accept","text/html, */*");
        xmlHttp.setRequestHeader("X-CSRF-TOKEN", getCookie("csrf_token"));
        sParam = getFormParamMultipleDataJSON(f);
        xmlHttp.send(sParam);
    }
    else
    {
        alert("Fail create xmlHttp");
    }
}

function doForm(objForm, url)
{
    var f = objForm;
    var sParam = '';
    var xmlHttp = createXmlHttp();
    if(xmlHttp != null)
    {
        xmlHttp.onreadystatechange = function()
        {
            if(xmlHttp.readyState == 4)
            {
                if(xmlHttp.status == 200)
                {
                    jump(url);
                }
            }
        };
        xmlHttp.open(f.method, f.action, true);
        xmlHttp.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
        //xmlHttp.setRequestHeader("User-Agent","MXL Browser 1.0");
        xmlHttp.setRequestHeader("Accept","text/html, */*");
        xmlHttp.setRequestHeader("Accept","text/html, */*");
        xmlHttp.setRequestHeader("X-CSRF-TOKEN", getCookie("csrf_token"));
        sParam = getFormParam(f);
        xmlHttp.send(sParam);
    }
    else
    {
        alert("Fail create xmlHttp");
    }
}
function enableAllHref(enable)
{
    var i = 0;
    var all = document.getElementsByTagName("a");
    for(i=0; i<all.length; i++)
    all[i].onclick = function(){return (enable==0?false:true);};
}
function createMenu()
{
    _w("<div id='leftNav'>");
    showLogo();
    _w("<ul id='menuList'>");
	
    _w("<li>Status</li>");
    _w("<ul id='subMenuStatus'>");
    _w("    <li><a href='devStatus.html'>Device Status</a></li>");
    _w("    <li><a href='phyRates.html'>MoCA Link Rates</a></li>");
    _w("</ul>");
	
    _w("<li>Settings</li>");
    _w("<ul id='subMenuSetup'>");
    _w("    <li><a href='index.html'>MoCA settings</a></li>");
    _w("    <li><a href='devSetup.html'>Device settings</a></li>");
    _w("    <li><a href='security.html'>Security settings</a></li>");
    _w("</ul>");
	

    _w("<li>Advanced</li>");
    _w("<ul id='subMenuAdvanced'>");
    _w("    <li><a href='upgrade.html'>Upgrade</a></li>");
    _w("    <li><a href='restore.html'>Restore</a></li>");
    _w("    <li><a href='reboot.html'>Reboot</a></li>");
    _w("</ul>");
	
    _w("</ul>");
	
    _w("</div>");
//    enableAllHref(0);
}

function showLogo()
{
    _w("<div id='logo'>");
    _w("<table align='left'>");
    _w("<tr><td><img alt='goCoax' src='logo.png' id='logoPic'></td></tr>");
    _w("</table>");
	
    _w("</div>");
}

function showRight(whichPage)
{
    var page = whichPage.toLowerCase();
	
    _w('<div id="rightHelp">');
    _w('<div id="helpInfo">');
    _w("<table>");
    if(page == "mocasetup")
    {
        _w("<tr><td>This screen is the first screen you will see when accessing the Coax bridge. Most users will be able to configure the bridge and get it working properly using only the settings on this screen. Pick the bands you'd like to scan. The scan offset is the offset in 25MHz steps starting from 0 MHz for the scan mask. The scan mask defines the channels to be scanned. The channel represents the center frequency of the beacon. Tx Power can be used to adjust the TX power for RF interface,and the Preferred NC is related with MoCA spec.Click the button Reboot can reboot the system,Click the button Restores Defaults can restore the system to factory default values.</td></tr>");
    }
    else if(page == "devsetup")
    {
        _w("<tr><td>This screen allows you to configure the IP mode and telnet server.Select 'DHCP automatic configuration' if your network has a DHCP server. If you choose Static IP address, you must configure the IP address for each coax bridge (note that each IP address must be unique. The new IP address will be used only after reset).Select 'c.Link Local automatic configuration' if there are no DHCP server in this network and you want make zero config for the IP.The IP address will not apply if Automatic Configuration (DHCP) is selected.If you enable MoCA telnet, then you can access the bridge by telnet protocol.</td></tr>");
    }
    else if(page == "security")
    {
        _w("<tr><td>This screen allows you to change the admin password for the bridge and the network security password for the Coax network. It is strongly recommended that you change the factory default password, the default admin password is maxlinear and the default network password is 99999999988888888. All users who try to access the bridge will be prompted for the bridge's password. The new admin password must not exceed 15 characters in length and must not include any spaces. The new network security password must be 12~17 digits. </td></tr>");
    }
    else if(page == "devstatus")
    {
        _w("<tr><td>This screen displays the current firmware version. Firmware should only be upgraded if you experience problems with the bride. Also displays the current IP address and MAC address of the bridge.The link status,node version and MoCA network version are displayed here.</td></tr>");
    }
    else if(page == "phyrates")
    {
        _w("<tr><td>This screen displays the current link status (PHY rate in Mbps) of each coax bridge relative to other nodes on the coax network.This data rate is an average of the Tx and Rx data rates between bridges.</td></tr>");
    }
    else if(page == "upgrade")
    {
        _w("<tr><td>You must be very careful when upgrade firmware,it may damage your device and can not work.you should following the step and do not remove power.</td></tr>");
    }
    else if(page == "reboot")
    {
        _w("<tr><td>Reboot may take about 10 seconds.<br />When rebooting this page will count down for 10 seconds,<br />And it will try connect to index page automatically.<br />Please refresh this page or input the correct URL address     manually if it is failed to connect with index page.</td></tr>");
    }
	
    _w("</table>");
    _w("</div>");
    _w("</div>");
}

function showFooter()
{
    _w("<div id='footer'>");
	
    _w("<table align='center'>");
    _w("    <tr><td><a href='mailto:support@gocoax.com' target='_blank'>Contact goCoax Customer Support</a></td></tr>");
    _w("</table>");
	
    _w("</div>");
    setTimeout("enableAllHref(1)", 3000);
}

function checkLenLimit(s, min, max)
{
    if(s.length < min || s.length > max)
    {
        return false;
    }

    return true;
}

function checkAllDigital(s)
{
    var reg = /\D/;
    return (s.match(reg) == null);
}

function checkIpAddr(ip)
{
    var reg = /^(([0-9]{1,3}\.){3}[0-9]{1,3})/;
    return (ip.match(reg) != null);
}


function checkForHexValue(str)
{
    var hexVal = "0x";    
    var hexStr = str.substr(0,2);
    var lchexStr = hexStr.toLocaleLowerCase();

    return lchexStr.localeCompare(hexVal);
}

function split64Hex(inStr)
{
    var result;

    if (inStr.length < 11)
    {
        var hiLen = 0;
        var loLen = inStr.length - 2;

        var hiNibble = 0;
        var loNibble = parseInt( inStr.substring(2,2+loLen) , 16 ) ;
    }
    else
    {
        var hiLen = inStr.length - 10;
        var loLen = 8;

        var hiNibble = parseInt( inStr.substring(2,2+hiLen) , 16 ) ;
        var loNibble = parseInt( inStr.substring(2+hiLen,2+hiLen+loLen) , 16 ) ;
    }
    
    result = hiNibble + " " + loNibble

    return result;
}

function ip2num(inDot)
{
    var tmp = inDot.split('.');
    return ((((((+tmp[0])*256)+(+tmp[1]))*256)+(+tmp[2]))*256)+(+tmp[3]);
}


function str2unicode(inStr)
{
    var retVal = '';

    for(i=0; i<inStr.length; i++)
    {
        intVal = inStr.charCodeAt(i);
        retVal += parseInt(intVal,10).toString(16);
    }

    return retVal;
}
function isAscii(c){return (c>0 && c<0x80);}
function hex2ascii(arguments)
{
    var i=0, j=0, b=0, hex=0, s="";
    for(j=0;j<arguments.length;j++)
    {
        hex = parseInt(arguments[j],16);
        for(i=0;i<4;i++)
        {
            b=(hex>>((3-i)*8)) & 0xff;
            if(!isAscii(b))return s;
            s += String.fromCharCode(b);
        }
    }
    return s;
}

function byte2ascii(arguments)
{
    var i=0, b=0, hex=0, s="";
    hex = parseInt(arguments,16);
    for(i=0;i<4;i++)
    {
        b=(hex>>((3-i)*8)) & 0xff;
        if(!isAscii(b))return s;
        s += String.fromCharCode(b);
    }
    return s;
}

function sha256(ascii) {
  function rightRotate(value, amount) {
    return (value>>>amount) | (value<<(32 - amount));
  };
  
  var mathPow = Math.pow;
  var maxWord = mathPow(2, 32);
  var lengthProperty = 'length'
  var i, j; // Used as a counter across the whole file
  var result = ''

  var words = [];
  var asciiBitLength = ascii[lengthProperty]*8;
  
  //* caching results is optional - remove/add slash from front of this line to toggle
  // Initial hash value: first 32 bits of the fractional parts of the square roots of the first 8 primes
  // (we actually calculate the first 64, but extra values are just ignored)
  var hash = sha256.h = sha256.h || [];
  // Round constants: first 32 bits of the fractional parts of the cube roots of the first 64 primes
  var k = sha256.k = sha256.k || [];
  var primeCounter = k[lengthProperty];
  /*/
  var hash = [], k = [];
  var primeCounter = 0;
  //*/

  var isComposite = {};
  for (var candidate = 2; primeCounter < 64; candidate++) {
    if (!isComposite[candidate]) {
      for (i = 0; i < 313; i += candidate) {
        isComposite[i] = candidate;
      }
      hash[primeCounter] = (mathPow(candidate, .5)*maxWord)|0;
      k[primeCounter++] = (mathPow(candidate, 1/3)*maxWord)|0;
    }
  }
  
  ascii += '\x80' // Append Ƈ' bit (plus zero padding)
  while (ascii[lengthProperty]%64 - 56) ascii += '\x00' // More zero padding
  for (i = 0; i < ascii[lengthProperty]; i++) {
    j = ascii.charCodeAt(i);
    if (j>>8) return; // ASCII check: only accept characters in range 0-255
    words[i>>2] |= j << ((3 - i)%4)*8;
  }
  words[words[lengthProperty]] = ((asciiBitLength/maxWord)|0);
  words[words[lengthProperty]] = (asciiBitLength)
  
  // process each chunk
  for (j = 0; j < words[lengthProperty];) {
    var w = words.slice(j, j += 16); // The message is expanded into 64 words as part of the iteration
    var oldHash = hash;
    // This is now the undefinedworking hash", often labelled as variables a...g
    // (we have to truncate as well, otherwise extra entries at the end accumulate
    hash = hash.slice(0, 8);
    
    for (i = 0; i < 64; i++) {
      var i2 = i + j;
      // Expand the message into 64 words
      // Used below if 
      var w15 = w[i - 15], w2 = w[i - 2];

      // Iterate
      var a = hash[0], e = hash[4];
      var temp1 = hash[7]
        + (rightRotate(e, 6) ^ rightRotate(e, 11) ^ rightRotate(e, 25)) // S1
        + ((e&hash[5])^((~e)&hash[6])) // ch
        + k[i]
        // Expand the message schedule if needed
        + (w[i] = (i < 16) ? w[i] : (
            w[i - 16]
            + (rightRotate(w15, 7) ^ rightRotate(w15, 18) ^ (w15>>>3)) // s0
            + w[i - 7]
            + (rightRotate(w2, 17) ^ rightRotate(w2, 19) ^ (w2>>>10)) // s1
          )|0
        );
      // This is only used once, so *could* be moved below, but it only saves 4 bytes and makes things unreadble
      var temp2 = (rightRotate(a, 2) ^ rightRotate(a, 13) ^ rightRotate(a, 22)) // S0
        + ((a&hash[1])^(a&hash[2])^(hash[1]&hash[2])); // maj
      
      hash = [(temp1 + temp2)|0].concat(hash); // We don't bother trimming off the extra ones, they're harmless as long as we're truncating when we do the slice()
      hash[4] = (hash[4] + temp1)|0;
    }
    
    for (i = 0; i < 8; i++) {
      hash[i] = (hash[i] + oldHash[i])|0;
    }
  }
  
  for (i = 0; i < 8; i++) {
    for (j = 3; j + 1; j--) {
      var b = (hash[i]>>(j*8))&255;
      result += ((b < 16) ? 0 : '') + b.toString(16);
    }
  }
  return result;
};

function checkOldPwd(oldPwd, oldAdminPassVal)
{
    console.log(sha256(oldPwd).toLowerCase());
    console.log(oldAdminPassVal.toLowerCase());
    return ((oldPwd == oldAdminPassVal) || (oldAdminPassVal.toLowerCase() == sha256(oldPwd).toLowerCase()));
}

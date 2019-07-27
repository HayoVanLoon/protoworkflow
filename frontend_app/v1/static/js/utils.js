function doHttpGet(url, callback, errorCb) {
  let httpReq = new XMLHttpRequest();
  httpReq.onreadystatechange = function () {
    if (httpReq.readyState === 4 && httpReq.status < 300) {
      callback(httpReq.responseText);
    } else if (typeof (errorCb) !== typeof (Function)) {
      console.log(httpReq.responseText);
    } else if (httpReq.readyState === 4 && httpReq.status >= 400) {
      errorCb(httpReq.responseText);
    }
  };
  httpReq.open("GET", url, true);
  httpReq.send(null);
}

function doHttpPostForm(url, body, callback, errorCb) {
  let httpReq = new XMLHttpRequest();
  httpReq.onreadystatechange = function () {
    if (httpReq.readyState === 4 && httpReq.status < 300) {
      callback(httpReq.responseText);
    } else if (typeof (errorCb) === typeof (Function) && httpReq.readyState === 4 && httpReq.status >= 400) {
      errorCb(httpReq.responseText);
    }
  };
  httpReq.open("POST", url, true);
  httpReq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  httpReq.send(body);
}

function sendHttpReq(method, url, body, callback, errorCb, contentType, headers) {
  let httpReq = new XMLHttpRequest();

  if (!contentType) {
    contentType = 'application/x-www-form-urlencoded';
  }
  if (!headers) {
    headers = [];
  }
  if (typeof (errorCb) !== typeof (Function)) {
    errorCb = console.log;
  }

  httpReq.onreadystatechange = function () {
    if (httpReq.readyState === 4 && httpReq.status < 300) {
      callback(httpReq.responseText);
    } else if (httpReq.readyState === 4 && httpReq.status >= 400) {
      errorCb(httpReq.responseText);
    }
  };

  httpReq.open(method, url, true);

  for (let i = 0; i < headers.length; i += 1) {
    httpReq.setRequestHeader(headers[i][0], headers[i][1]);
  }

  switch (method) {
    case 'HEAD':
    case 'GET':
    case 'DELETE':
      httpReq.send();
      break;
    default:
      httpReq.setRequestHeader("Content-type", contentType);
      httpReq.send(body);
  }
}

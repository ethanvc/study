
function init() {
    const filter: chrome.webRequest.RequestFilter = { urls: ['<all_urls>'] };
    chrome.webRequest.onBeforeRequest.addListener(onBeforeRequest, filter);
    chrome.webRequest.onCompleted.addListener(onCompleted, filter);
    chrome.webRequest.onErrorOccurred.addListener(onErrorOccurred, filter);
}

function onBeforeRequest(details: chrome.webRequest.WebRequestBodyDetails): void {
    console.log(`before request, id:${details.requestId}, url: ${typeof details.url}`);
}

function onCompleted(details: chrome.webRequest.WebResponseCacheDetails): void {
    console.log(`completed request, id:${details.requestId}, url: ${details.url}`);
}

function onErrorOccurred(details: chrome.webRequest.WebResponseErrorDetails): void {
    console.log(`error occurred, id:${details.requestId}, url: ${details.url}`);
}

export { init };
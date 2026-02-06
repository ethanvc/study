import { NetEvaluator, RequestStatus } from "./net_evaluator";

function init() {
    const filter: chrome.webRequest.RequestFilter = { urls: ['<all_urls>'] };
    chrome.webRequest.onBeforeRequest.addListener(onBeforeRequest, filter);
    chrome.webRequest.onCompleted.addListener(onCompleted, filter);
    chrome.webRequest.onErrorOccurred.addListener(onErrorOccurred, filter);
}

const netEvaluator = new NetEvaluator();

function onBeforeRequest(details: chrome.webRequest.WebRequestBodyDetails): void {
    netEvaluator.addRequest(details.tabId, details.url, details.requestId);
}

function onCompleted(details: chrome.webRequest.WebResponseCacheDetails): void {
    netEvaluator.finishRequest(details.requestId, RequestStatus.Ok);
}

function onErrorOccurred(details: chrome.webRequest.WebResponseErrorDetails): void {
    netEvaluator.finishRequest(details.requestId, RequestStatus.Failed);
}

export { init };
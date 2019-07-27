/*
 * Copyright 2019 Hayo van Loon
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

(function () {
  const contactUrl = "/contact";
  const storageStatsUrl = '/storage-stats';
  const questionUrl = '/question';
  const complaintUrl = '/complaint';
  const feedbackUrl = '/feedback';

  const messageInputElem = document.getElementById("message-input");
  const senderInputElem = document.getElementById("sender-input");
  const productIdInputElem = document.getElementById("product-id-input");
  const messageSubmitSpanElem = document.getElementById("message-submit-span");

  const storageStatsSpan = document.getElementById('storage-stats-span');
  const questionSpan = document.getElementById('question-span');
  const complaintSpan = document.getElementById('complaint-span');
  const feedbackSpan = document.getElementById('feedback-span');

  function createMessage(message, sender, productId) {
    let body = 'message=' + encodeURI(message) +
        '&sender=' + encodeURI(sender);
    if (!!productId) {
      body += '&product-id=' + encodeURI(productId);
    }
    function cb(body) {
      messageSubmitSpanElem.innerText = body;
    }
    function errorCb(body) {
      messageSubmitSpanElem.innerText = body
    }
    sendHttpReq('POST', contactUrl, body, cb, errorCb);
  }

  function submitMessage() {
    let ok = messageInputElem.value.trim() !== '' && senderInputElem.value.trim() !== '';
    if (!ok) {
      messageSubmitSpanElem.innerText = 'You must specify at least message and sender.'
    } else {
      messageSubmitSpanElem.innerText = '';
      createMessage(messageInputElem.value, senderInputElem.value, productIdInputElem.value);
    }
  }

  function clickButtonFn(url, spanElem) {
    return function() {
      doHttpGet(url, function(body){spanElem.innerText = body;}, console.log)
    };
  }

  document.getElementById('message-submit').addEventListener('click', submitMessage);
  document.getElementById('storage-stats-btn')
      .addEventListener('click', clickButtonFn(storageStatsUrl, storageStatsSpan));
  document.getElementById('question-btn')
      .addEventListener('click', clickButtonFn(questionUrl, questionSpan));
  document.getElementById('complaint-btn')
      .addEventListener('click', clickButtonFn(complaintUrl, complaintSpan));
  document.getElementById('feedback-btn')
      .addEventListener('click', clickButtonFn(feedbackUrl, feedbackSpan));
})();
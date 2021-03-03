const highlightMsg = (function () {
  let prevId = '';

  return (hash) => {
    if (hash === '') return;
    const id = location.hash.substring(1);
    const $msg = $('#' + CSS.escape(id));
    // open thread if inner message is specified by URL fragment
    $msg.closest('details.slacklog-thread').attr('open', '');
    // begin highlight animation (see assets/css/slacklog.css)
    if (prevId !== '') {
      $('#' + CSS.escape(prevId)).removeClass('slacklog-highlight');
    }
    $msg.addClass('slacklog-highlight');
    prevId = id;
  };
})();

window.addEventListener('hashchange', () => {
  highlightMsg(location.hash);
});
window.addEventListener('DOMContentLoaded', () => {
  highlightMsg(location.hash);
});

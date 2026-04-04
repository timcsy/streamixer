/**
 * Streamixer HLS 播放器
 */
(function() {
  'use strict';

  function initPlayer(container) {
    var url = container.getAttribute('data-hls-url');
    var video = container.querySelector('.streamixer-video');
    var errorEl = container.querySelector('.streamixer-error');

    if (!url || !video) return;

    function showError(msg) {
      if (errorEl) {
        errorEl.textContent = msg || '無法載入串流，請稍後再試。';
        errorEl.style.display = 'block';
      }
      video.style.display = 'none';
    }

    if (typeof Hls !== 'undefined' && Hls.isSupported()) {
      var hls = new Hls({
        maxBufferLength: 30,
        maxMaxBufferLength: 60,
      });
      hls.loadSource(url);
      hls.attachMedia(video);
      hls.on(Hls.Events.MANIFEST_PARSED, function() {
        // 就緒，使用者可手動播放
      });
      hls.on(Hls.Events.ERROR, function(event, data) {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              showError('網路錯誤，無法載入串流。');
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              hls.recoverMediaError();
              break;
            default:
              showError('播放發生錯誤。');
              hls.destroy();
              break;
          }
        }
      });
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      // Safari 原生支援
      video.src = url;
      video.addEventListener('error', function() {
        showError('無法載入串流。');
      });
    } else {
      showError('您的瀏覽器不支援 HLS 播放。');
    }
  }

  function initAll() {
    var containers = document.querySelectorAll('.streamixer-player-container');
    containers.forEach(initPlayer);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initAll);
  } else {
    initAll();
  }
})();

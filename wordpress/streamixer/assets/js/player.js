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

  function initDownloadButton(toolbar) {
    var btn = toolbar.querySelector('.streamixer-download-btn');
    var progress = toolbar.querySelector('.streamixer-download-progress');
    var bar = toolbar.querySelector('.streamixer-download-progress-bar span');
    var text = toolbar.querySelector('.streamixer-download-progress-text');
    var downloadUrl = toolbar.getAttribute('data-download-url');
    var progressUrl = toolbar.getAttribute('data-progress-url');
    if (!btn || !downloadUrl || !progressUrl) return;

    var polling = false;

    function setProgress(percent, label) {
      if (bar) bar.style.width = percent + '%';
      if (text) text.textContent = label;
    }

    function triggerDownload() {
      var a = document.createElement('a');
      a.href = downloadUrl;
      a.download = '';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }

    function poll() {
      fetch(progressUrl, { cache: 'no-store' })
        .then(function(r) { return r.json(); })
        .then(function(data) {
          if (data.status === 'failed') {
            setProgress(0, '合成失敗，請稍後再試');
            btn.disabled = false;
            polling = false;
            return;
          }
          if (data.ready || data.status === 'completed') {
            setProgress(100, '下載中…');
            polling = false;
            triggerDownload();
            setTimeout(function() {
              progress.style.display = 'none';
              btn.disabled = false;
              setProgress(0, '準備中…');
            }, 2000);
            return;
          }
          setProgress(data.percent || 0, '影片合成中 ' + (data.percent || 0) + '%（' + (data.done || 0) + '/' + (data.total || 0) + ' 分段）');
          setTimeout(poll, 1000);
        })
        .catch(function() {
          setTimeout(poll, 2000);
        });
    }

    btn.addEventListener('click', function(e) {
      e.preventDefault();
      if (polling) return;
      polling = true;
      btn.disabled = true;
      progress.style.display = 'block';
      setProgress(0, '檢查進度…');
      poll();
    });
  }

  function initAll() {
    var containers = document.querySelectorAll('.streamixer-player-container');
    containers.forEach(initPlayer);
    var toolbars = document.querySelectorAll('.streamixer-player-toolbar');
    toolbars.forEach(initDownloadButton);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initAll);
  } else {
    initAll();
  }
})();

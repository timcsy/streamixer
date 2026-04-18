<?php
/**
 * 外掛設定頁
 */
class Streamixer_Settings {

	public static function add_menu() {
		add_options_page(
			'Streamixer 設定',
			'Streamixer',
			'manage_options',
			'streamixer-settings',
			array( __CLASS__, 'render_page' )
		);
	}

	public static function register_settings() {
		register_setting( 'streamixer_options', 'streamixer_service_url', array(
			'type'              => 'string',
			'sanitize_callback' => 'esc_url_raw',
			'default'           => 'http://localhost:8080',
		) );

		register_setting( 'streamixer_options', 'streamixer_public_url', array(
			'type'              => 'string',
			'sanitize_callback' => 'esc_url_raw',
			'default'           => '',
		) );

		register_setting( 'streamixer_options', 'streamixer_api_key', array(
			'type'              => 'string',
			'sanitize_callback' => 'sanitize_text_field',
			'default'           => '',
		) );

		register_setting( 'streamixer_options', 'streamixer_auto_cleanup', array(
			'type'              => 'string',
			'sanitize_callback' => 'sanitize_text_field',
			'default'           => '1',
		) );

		register_setting( 'streamixer_options', 'streamixer_default_background', array(
			'type'              => 'integer',
			'sanitize_callback' => 'absint',
			'default'           => 0,
		) );

		// 快取設定（同步到 Streamixer 後端）
		register_setting( 'streamixer_options', 'streamixer_cache_ttl', array(
			'type'              => 'string',
			'sanitize_callback' => 'sanitize_text_field',
			'default'           => '30m',
		) );
		register_setting( 'streamixer_options', 'streamixer_cache_sweep_interval', array(
			'type'              => 'string',
			'sanitize_callback' => 'sanitize_text_field',
			'default'           => '5m',
		) );
		register_setting( 'streamixer_options', 'streamixer_cache_max_size_mb', array(
			'type'              => 'integer',
			'sanitize_callback' => 'absint',
			'default'           => 0,
		) );

		add_settings_section(
			'streamixer_main',
			'服務設定',
			null,
			'streamixer-settings'
		);

		add_settings_field(
			'streamixer_service_url',
			'Streamixer 服務 URL',
			array( __CLASS__, 'render_url_field' ),
			'streamixer-settings',
			'streamixer_main'
		);

		add_settings_field(
			'streamixer_public_url',
			'前端播放 URL',
			array( __CLASS__, 'render_public_url_field' ),
			'streamixer-settings',
			'streamixer_main'
		);

		add_settings_field(
			'streamixer_api_key',
			'API Key',
			array( __CLASS__, 'render_api_key_field' ),
			'streamixer-settings',
			'streamixer_main'
		);

		add_settings_field(
			'streamixer_auto_cleanup',
			'同步後清除本地檔案',
			array( __CLASS__, 'render_auto_cleanup_field' ),
			'streamixer-settings',
			'streamixer_main'
		);

		add_settings_field(
			'streamixer_default_background',
			'預設背景圖片',
			array( __CLASS__, 'render_background_field' ),
			'streamixer-settings',
			'streamixer_main'
		);

		// 快取設定區塊
		add_settings_section(
			'streamixer_cache',
			'影片快取設定',
			array( __CLASS__, 'render_cache_section_intro' ),
			'streamixer-settings'
		);

		add_settings_field(
			'streamixer_cache_ttl',
			'快取過期時間 (TTL)',
			array( __CLASS__, 'render_cache_ttl_field' ),
			'streamixer-settings',
			'streamixer_cache'
		);

		add_settings_field(
			'streamixer_cache_sweep_interval',
			'清掃頻率',
			array( __CLASS__, 'render_cache_sweep_field' ),
			'streamixer-settings',
			'streamixer_cache'
		);

		add_settings_field(
			'streamixer_cache_max_size_mb',
			'快取容量上限 (MB)',
			array( __CLASS__, 'render_cache_max_size_field' ),
			'streamixer-settings',
			'streamixer_cache'
		);

		// 儲存後 push 到 Streamixer 後端
		add_action( 'update_option_streamixer_cache_ttl', array( __CLASS__, 'push_cache_config' ), 10, 0 );
		add_action( 'update_option_streamixer_cache_sweep_interval', array( __CLASS__, 'push_cache_config' ), 10, 0 );
		add_action( 'update_option_streamixer_cache_max_size_mb', array( __CLASS__, 'push_cache_config' ), 10, 0 );
	}

	public static function render_cache_section_intro() {
		$current = self::fetch_current_cache_config();
		echo '<p>這些設定會即時套用到 Streamixer 後端（不需重啟容器）。</p>';
		if ( $current ) {
			$usage_mb = round( $current['cache_usage_bytes'] / 1048576, 1 );
			echo '<p><strong>後端目前狀態：</strong>TTL ' . esc_html( $current['cache_ttl'] )
				. '、清掃頻率 ' . esc_html( $current['cache_sweep_interval'] )
				. '、容量上限 ' . ( $current['cache_max_size'] > 0 ? esc_html( round( $current['cache_max_size'] / 1048576 ) ) . ' MB' : '不限制' )
				. '、目前使用 ' . esc_html( $usage_mb ) . ' MB</p>';
		} else {
			echo '<p><em>無法取得後端狀態，請確認 Streamixer 服務連線正常。</em></p>';
		}
	}

	public static function render_cache_ttl_field() {
		$value = get_option( 'streamixer_cache_ttl', '30m' );
		?>
		<input type="text" name="streamixer_cache_ttl"
		       value="<?php echo esc_attr( $value ); ?>"
		       class="regular-text" placeholder="30m">
		<p class="description">素材無人存取多久後清除快取分段。格式：<code>30m</code>、<code>1h</code>、<code>2h30m</code>。至少 1 秒。</p>
		<?php
	}

	public static function render_cache_sweep_field() {
		$value = get_option( 'streamixer_cache_sweep_interval', '5m' );
		?>
		<input type="text" name="streamixer_cache_sweep_interval"
		       value="<?php echo esc_attr( $value ); ?>"
		       class="regular-text" placeholder="5m">
		<p class="description">背景清掃排程的執行間隔。格式同上。至少 10 秒。</p>
		<?php
	}

	public static function render_cache_max_size_field() {
		$value = get_option( 'streamixer_cache_max_size_mb', 0 );
		?>
		<input type="number" name="streamixer_cache_max_size_mb"
		       value="<?php echo esc_attr( $value ); ?>"
		       min="0" step="1" class="small-text"> MB
		<p class="description">tmpfs 快取容量上限。填 0 表示不限制（但仍受 tmpfs 本身大小限制）。超過上限時會以 LRU 淘汰最久未用的素材。</p>
		<?php
	}

	/**
	 * 將 WordPress 設定 push 至 Streamixer 後端 /config
	 */
	public static function push_cache_config() {
		$url = Streamixer_API::get_service_url() . '/config';
		$ttl       = get_option( 'streamixer_cache_ttl', '30m' );
		$sweep     = get_option( 'streamixer_cache_sweep_interval', '5m' );
		$mb        = (int) get_option( 'streamixer_cache_max_size_mb', 0 );
		$payload   = array(
			'cache_ttl'            => $ttl,
			'cache_sweep_interval' => $sweep,
			'cache_max_size'       => $mb * 1048576,
		);
		$headers   = array( 'Content-Type' => 'application/json' );
		$api_key = get_option( 'streamixer_api_key', '' );
		if ( ! empty( $api_key ) ) {
			$headers['X-API-Key'] = $api_key;
		}
		wp_remote_request( $url, array(
			'method'  => 'PUT',
			'timeout' => 10,
			'headers' => $headers,
			'body'    => wp_json_encode( $payload ),
		) );
	}

	/**
	 * 從 Streamixer 後端讀取目前快取設定
	 */
	public static function fetch_current_cache_config() {
		$url      = Streamixer_API::get_service_url() . '/config';
		$response = wp_remote_get( $url, array( 'timeout' => 5 ) );
		if ( is_wp_error( $response ) ) {
			return null;
		}
		$body = wp_remote_retrieve_body( $response );
		$data = json_decode( $body, true );
		if ( ! is_array( $data ) ) {
			return null;
		}
		return $data;
	}

	public static function render_url_field() {
		$value = get_option( 'streamixer_service_url', 'http://localhost:8080' );
		?>
		<input type="url" name="streamixer_service_url"
		       value="<?php echo esc_attr( $value ); ?>"
		       class="regular-text"
		       placeholder="http://localhost:8080">
		<p class="description">Streamixer 串流合成服務的 URL。</p>
		<?php
	}

	public static function render_public_url_field() {
		$value = get_option( 'streamixer_public_url', '' );
		?>
		<input type="url" name="streamixer_public_url"
		       value="<?php echo esc_attr( $value ); ?>"
		       class="regular-text"
		       placeholder="http://localhost:8081">
		<p class="description">瀏覽器端存取 Streamixer 的 URL。若留空則使用服務 URL。<br>
		Docker 環境中，服務 URL 通常是容器內部名稱（如 <code>http://streamixer:8080</code>），而前端播放 URL 是外部可存取的位址（如 <code>http://localhost:8081</code>）。</p>
		<?php
	}

	public static function render_api_key_field() {
		$value = get_option( 'streamixer_api_key', '' );
		?>
		<input type="text" name="streamixer_api_key"
		       value="<?php echo esc_attr( $value ); ?>"
		       class="regular-text"
		       placeholder="留空表示不需要認證">
		<p class="description">Streamixer 服務的 API Key。需與服務端的 <code>API_KEY</code> 環境變數一致。</p>
		<?php
	}

	public static function render_auto_cleanup_field() {
		$value = get_option( 'streamixer_auto_cleanup', '1' );
		?>
		<input type="hidden" name="streamixer_auto_cleanup" value="0">
		<label>
			<input type="checkbox" name="streamixer_auto_cleanup" value="1" <?php checked( $value, '1' ); ?>>
			素材同步至 Streamixer 後，自動刪除 WordPress 端的原始檔案以節省儲存空間
		</label>
		<p class="description">啟用後，音檔、圖片、字幕在同步成功後會從 WordPress 媒體庫中刪除實際檔案（保留記錄）。</p>
		<?php
	}

	public static function render_background_field() {
		wp_enqueue_media();
		$bg_id  = get_option( 'streamixer_default_background', 0 );
		$bg_url = $bg_id ? wp_get_attachment_url( $bg_id ) : '';
		?>
		<input type="hidden" name="streamixer_default_background" id="streamixer_default_bg_id"
		       value="<?php echo esc_attr( $bg_id ); ?>">
		<button type="button" class="button" id="streamixer_default_bg_btn">選擇圖片</button>
		<button type="button" class="button" id="streamixer_default_bg_clear">清除</button>
		<span id="streamixer_default_bg_preview" style="margin-left:10px; color:#666;">
			<?php echo $bg_url ? esc_html( basename( $bg_url ) ) : '未設定'; ?>
		</span>
		<p class="description">當素材組合未上傳背景圖片時使用的預設圖片。</p>

		<script>
		jQuery(document).ready(function($) {
			$('#streamixer_default_bg_btn').on('click', function(e) {
				e.preventDefault();
				var frame = wp.media({ multiple: false, library: { type: 'image' } });
				frame.on('select', function() {
					var attachment = frame.state().get('selection').first().toJSON();
					$('#streamixer_default_bg_id').val(attachment.id);
					$('#streamixer_default_bg_preview').text(attachment.filename);
				});
				frame.open();
			});
			$('#streamixer_default_bg_clear').on('click', function(e) {
				e.preventDefault();
				$('#streamixer_default_bg_id').val('');
				$('#streamixer_default_bg_preview').text('未設定');
			});
		});
		</script>
		<?php
	}

	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			return;
		}

		// 測試連線
		$connection_status = '';
		if ( isset( $_GET['test-connection'] ) ) {
			$url      = get_option( 'streamixer_service_url', 'http://localhost:8080' );
			$response = wp_remote_get( $url . '/health', array( 'timeout' => 5 ) );
			if ( is_wp_error( $response ) ) {
				$connection_status = '<div class="notice notice-error"><p>連線失敗：' . esc_html( $response->get_error_message() ) . '</p></div>';
			} else {
				$code = wp_remote_retrieve_response_code( $response );
				if ( 200 === $code ) {
					$connection_status = '<div class="notice notice-success"><p>✓ 連線成功！</p></div>';
				} else {
					$connection_status = '<div class="notice notice-error"><p>連線失敗（HTTP ' . $code . '）</p></div>';
				}
			}
		}
		?>
		<div class="wrap">
			<h1>Streamixer 設定</h1>
			<?php echo $connection_status; ?>
			<?php settings_errors(); ?>
			<form method="post" action="options.php">
				<?php
				settings_fields( 'streamixer_options' );
				do_settings_sections( 'streamixer-settings' );
				submit_button();
				?>
			</form>
			<p>
				<a href="<?php echo admin_url( 'options-general.php?page=streamixer-settings&test-connection=1' ); ?>"
				   class="button">測試連線</a>
			</p>
		</div>
		<?php
	}
}

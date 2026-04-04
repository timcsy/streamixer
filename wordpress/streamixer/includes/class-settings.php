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

		register_setting( 'streamixer_options', 'streamixer_default_background', array(
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
			'streamixer_default_background',
			'預設背景圖片',
			array( __CLASS__, 'render_background_field' ),
			'streamixer-settings',
			'streamixer_main'
		);
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

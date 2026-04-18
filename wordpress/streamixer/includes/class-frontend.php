<?php
/**
 * 前端 Asset 管理與模板載入
 */
class Streamixer_Frontend {

	public static function init() {
		add_action( 'wp_enqueue_scripts', array( __CLASS__, 'register_assets' ) );
		add_filter( 'single_template', array( __CLASS__, 'single_template' ) );
		add_filter( 'archive_template', array( __CLASS__, 'archive_template' ) );
	}

	public static function register_assets() {
		wp_register_script(
			'hls-js',
			'https://cdn.jsdelivr.net/npm/hls.js@latest',
			array(),
			null,
			true
		);

		wp_register_script(
			'streamixer-player',
			STREAMIXER_PLUGIN_URL . 'assets/js/player.js',
			array( 'hls-js' ),
			STREAMIXER_VERSION,
			true
		);

		wp_register_style(
			'streamixer-player',
			STREAMIXER_PLUGIN_URL . 'assets/css/player.css',
			array(),
			STREAMIXER_VERSION
		);

		// 在素材組合頁面自動載入
		if ( is_singular( 'streamixer' ) || is_post_type_archive( 'streamixer' ) ) {
			wp_enqueue_script( 'streamixer-player' );
			wp_enqueue_style( 'streamixer-player' );
		}
	}

	/**
	 * 載入自訂 single 模板
	 */
	public static function single_template( $template ) {
		if ( is_singular( 'streamixer' ) ) {
			$plugin_template = STREAMIXER_PLUGIN_DIR . 'templates/single-streamixer.php';
			if ( file_exists( $plugin_template ) ) {
				return $plugin_template;
			}
		}
		return $template;
	}

	/**
	 * 載入自訂 archive 模板
	 */
	public static function archive_template( $template ) {
		if ( is_post_type_archive( 'streamixer' ) ) {
			$plugin_template = STREAMIXER_PLUGIN_DIR . 'templates/archive-streamixer.php';
			if ( file_exists( $plugin_template ) ) {
				return $plugin_template;
			}
		}
		return $template;
	}

	/**
	 * 渲染播放器 HTML（共用於 Shortcode、Block、模板）
	 */
	public static function render_player( $post_id ) {
		$stream_url = Streamixer_API::get_stream_url( $post_id );
		$title      = get_the_title( $post_id );

		wp_enqueue_script( 'streamixer-player' );
		wp_enqueue_style( 'streamixer-player' );

		ob_start();
		?>
		<?php
		$download_url    = Streamixer_API::get_download_url( $post_id );
		$progress_url    = Streamixer_API::get_progress_url( $post_id );
		$audio_url       = Streamixer_API::get_audio_url( $post_id );
		$transcript_url  = Streamixer_API::get_transcript_url( $post_id );
		$has_transcript  = Streamixer_API::has_transcript( $post_id );
		?>
		<div class="streamixer-player-container" data-hls-url="<?php echo esc_attr( $stream_url ); ?>">
			<video class="streamixer-video" controls playsinline>
				您的瀏覽器不支援影片播放。
			</video>
			<div class="streamixer-error" style="display:none;">
				無法載入串流，請稍後再試。
			</div>
		</div>
		<div class="streamixer-player-toolbar"
			data-download-url="<?php echo esc_attr( $download_url ); ?>"
			data-progress-url="<?php echo esc_attr( $progress_url ); ?>">
			<button type="button" class="streamixer-download-btn">⬇ 下載影片</button>
			<a href="<?php echo esc_attr( $audio_url ); ?>" class="streamixer-download-btn streamixer-download-link" download>⬇ 下載音檔</a>
			<?php if ( $has_transcript ) : ?>
			<a href="<?php echo esc_attr( $transcript_url ); ?>" class="streamixer-download-btn streamixer-download-link" download>⬇ 下載逐字稿</a>
			<?php endif; ?>
			<div class="streamixer-download-progress" style="display:none;">
				<div class="streamixer-download-progress-bar"><span></span></div>
				<div class="streamixer-download-progress-text">準備中…</div>
			</div>
		</div>
		<?php
		return ob_get_clean();
	}
}

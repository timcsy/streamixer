<?php
/**
 * Plugin Name: Streamixer
 * Description: 將音檔、背景圖片與字幕即時合成為 HLS 影片串流。管理素材組合並在前台播放。
 * Version: 1.0.0
 * Author: Streamixer
 * Text Domain: streamixer
 * Requires at least: 6.0
 * Requires PHP: 8.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

define( 'STREAMIXER_VERSION', '1.0.0' );
define( 'STREAMIXER_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
define( 'STREAMIXER_PLUGIN_URL', plugin_dir_url( __FILE__ ) );

// 載入 class 檔案
require_once STREAMIXER_PLUGIN_DIR . 'includes/class-cpt.php';
require_once STREAMIXER_PLUGIN_DIR . 'includes/class-settings.php';
require_once STREAMIXER_PLUGIN_DIR . 'includes/class-api.php';
require_once STREAMIXER_PLUGIN_DIR . 'includes/class-shortcode.php';
require_once STREAMIXER_PLUGIN_DIR . 'includes/class-frontend.php';

// 初始化
add_action( 'init', array( 'Streamixer_CPT', 'register' ) );
add_action( 'admin_menu', array( 'Streamixer_Settings', 'add_menu' ) );
add_action( 'admin_init', array( 'Streamixer_Settings', 'register_settings' ) );

// Shortcode
Streamixer_Shortcode::init();

// 前端 assets
Streamixer_Frontend::init();

// 允許上傳字幕檔案格式
add_filter( 'upload_mimes', function( $mimes ) {
	$mimes['srt'] = 'text/plain';
	$mimes['vtt'] = 'text/vtt';
	return $mimes;
} );

// 繞過 WordPress 的 filetype 驗證（srt/vtt 不在 WordPress 內建白名單中）
add_filter( 'wp_check_filetype_and_ext', function( $data, $file, $filename, $mimes ) {
	$ext = pathinfo( $filename, PATHINFO_EXTENSION );
	if ( in_array( strtolower( $ext ), array( 'srt', 'vtt' ), true ) ) {
		$data['ext']  = $ext;
		$data['type'] = ( 'vtt' === $ext ) ? 'text/vtt' : 'text/plain';
		$data['proper_filename'] = $filename;
	}
	return $data;
}, 10, 4 );

// Gutenberg Block（純 JS，不需 build）
add_action( 'init', 'streamixer_register_block' );
function streamixer_register_block() {
	register_block_type( 'streamixer/player', array(
		'render_callback' => 'streamixer_block_render',
		'attributes'      => array(
			'compositionId' => array( 'type' => 'number', 'default' => 0 ),
			'compositionTitle' => array( 'type' => 'string', 'default' => '' ),
		),
	) );
}

// 在 Gutenberg 編輯器載入 block JS
add_action( 'admin_enqueue_scripts', function( $hook ) {
	if ( ! in_array( $hook, array( 'post.php', 'post-new.php' ), true ) ) {
		return;
	}
	wp_enqueue_script(
		'streamixer-block-editor',
		STREAMIXER_PLUGIN_URL . 'assets/js/block.js',
		array( 'wp-blocks', 'wp-element', 'wp-components', 'wp-block-editor', 'wp-api-fetch' ),
		STREAMIXER_VERSION,
		true
	);
} );

function streamixer_block_render( $attributes ) {
	$id = isset( $attributes['compositionId'] ) ? intval( $attributes['compositionId'] ) : 0;
	if ( ! $id ) {
		return '<p class="streamixer-error-msg">Streamixer：未選擇素材組合。</p>';
	}
	$post = get_post( $id );
	if ( ! $post || 'streamixer' !== $post->post_type ) {
		return '<p class="streamixer-error-msg">Streamixer：找不到指定的素材。</p>';
	}
	return Streamixer_Frontend::render_player( $id );
}

// 讓 streamixer CPT 支援 REST API（Gutenberg Block 需要查詢素材列表）
add_action( 'init', function() {
	global $wp_post_types;
	if ( isset( $wp_post_types['streamixer'] ) ) {
		$wp_post_types['streamixer']->show_in_rest = true;
		$wp_post_types['streamixer']->rest_base = 'streamixer';
	}
}, 20 );

// 在文章編輯器加入「插入 Streamixer」按鈕（傳統編輯器）
add_action( 'media_buttons', 'streamixer_add_media_button' );
add_action( 'admin_footer', 'streamixer_media_button_modal' );

function streamixer_add_media_button() {
	$screen = get_current_screen();
	if ( $screen && 'streamixer' === $screen->post_type ) {
		return; // 素材組合自己的編輯頁不需要此按鈕
	}
	echo '<button type="button" class="button streamixer-insert-btn" style="padding-left:8px">';
	echo '<span class="dashicons dashicons-format-audio" style="vertical-align:text-bottom;margin-right:2px"></span> ';
	echo '插入 Streamixer</button>';
}

function streamixer_media_button_modal() {
	$screen = get_current_screen();
	if ( ! $screen || ! in_array( $screen->base, array( 'post', 'page' ), true ) ) {
		return;
	}
	if ( 'streamixer' === $screen->post_type ) {
		return;
	}

	$compositions = get_posts( array(
		'post_type'      => 'streamixer',
		'post_status'    => 'publish',
		'posts_per_page' => -1,
		'orderby'        => 'date',
		'order'          => 'DESC',
	) );
	?>
	<div id="streamixer-modal" style="display:none; position:fixed; top:0; left:0; right:0; bottom:0; background:rgba(0,0,0,0.6); z-index:100100; align-items:center; justify-content:center;">
		<div style="background:#fff; border-radius:8px; padding:24px; max-width:480px; width:90%; max-height:70vh; overflow-y:auto; box-shadow:0 4px 20px rgba(0,0,0,0.3);">
			<div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:16px;">
				<h2 style="margin:0; font-size:18px;">插入 Streamixer 播放器</h2>
				<button type="button" id="streamixer-modal-close" style="background:none; border:none; font-size:20px; cursor:pointer; color:#666;">&times;</button>
			</div>
			<?php if ( empty( $compositions ) ) : ?>
				<p style="color:#666;">尚無已發佈的素材組合。<a href="<?php echo admin_url( 'post-new.php?post_type=streamixer' ); ?>">前往新增</a></p>
			<?php else : ?>
				<input type="text" id="streamixer-search" placeholder="搜尋素材..." style="width:100%; padding:8px 12px; border:1px solid #ccc; border-radius:4px; margin-bottom:12px; box-sizing:border-box;">
				<div id="streamixer-list" style="border:1px solid #eee; border-radius:4px; max-height:40vh; overflow-y:auto;">
					<?php foreach ( $compositions as $comp ) : ?>
						<div class="streamixer-modal-item" data-slug="<?php echo esc_attr( $comp->post_name ); ?>" data-title="<?php echo esc_attr( $comp->post_title ); ?>"
						     style="padding:10px 14px; cursor:pointer; border-bottom:1px solid #f0f0f0; transition:background 0.15s;"
						     onmouseover="this.style.background='#f0f6fc'" onmouseout="this.style.background='transparent'">
							<div style="font-weight:500;"><?php echo esc_html( $comp->post_title ); ?></div>
							<div style="font-size:12px; color:#999; margin-top:2px;"><?php echo get_the_date( 'Y-m-d', $comp ); ?> · <code style="font-size:11px;">[streamixer id="<?php echo esc_attr( $comp->post_name ); ?>"]</code></div>
						</div>
					<?php endforeach; ?>
				</div>
			<?php endif; ?>
		</div>
	</div>
	<script>
	(function() {
		var modal = document.getElementById('streamixer-modal');
		if (!modal) return;

		// 開啟
		document.querySelectorAll('.streamixer-insert-btn').forEach(function(btn) {
			btn.addEventListener('click', function(e) {
				e.preventDefault();
				modal.style.display = 'flex';
			});
		});

		// 關閉
		document.getElementById('streamixer-modal-close').addEventListener('click', function() {
			modal.style.display = 'none';
		});
		modal.addEventListener('click', function(e) {
			if (e.target === modal) modal.style.display = 'none';
		});

		// 搜尋
		var search = document.getElementById('streamixer-search');
		if (search) {
			search.addEventListener('input', function() {
				var q = this.value.toLowerCase();
				document.querySelectorAll('.streamixer-modal-item').forEach(function(item) {
					var title = item.getAttribute('data-title').toLowerCase();
					item.style.display = title.indexOf(q) >= 0 ? '' : 'none';
				});
			});
		}

		// 選取插入
		document.querySelectorAll('.streamixer-modal-item').forEach(function(item) {
			item.addEventListener('click', function() {
				var slug = this.getAttribute('data-slug');
				var shortcode = '[streamixer id="' + slug + '"]';
				if (typeof tinymce !== 'undefined' && tinymce.activeEditor && !tinymce.activeEditor.isHidden()) {
					tinymce.activeEditor.execCommand('mceInsertContent', false, shortcode);
				} else {
					var textarea = document.getElementById('content');
					if (textarea) {
						var start = textarea.selectionStart;
						var text = textarea.value;
						textarea.value = text.substring(0, start) + shortcode + text.substring(textarea.selectionEnd);
						textarea.selectionStart = textarea.selectionEnd = start + shortcode.length;
						textarea.focus();
					}
				}
				modal.style.display = 'none';
			});
		});
	})();
	</script>
	<?php
}

// 啟用/停用 hook
register_activation_hook( __FILE__, function() {
	Streamixer_CPT::register();
	flush_rewrite_rules();
} );

register_deactivation_hook( __FILE__, function() {
	flush_rewrite_rules();
} );

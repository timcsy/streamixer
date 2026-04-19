<?php
/**
 * Custom Post Type 與 Meta Box
 */
class Streamixer_CPT {

	public static function register() {
		// Slug 可透過 filter 覆寫（避免與反向代理路徑如 /streamixer 衝突）
		// 範例：add_filter( 'streamixer_cpt_slug', fn() => 'my-media' );
		$cpt_slug = apply_filters( 'streamixer_cpt_slug', 'media' );
		$cat_slug = apply_filters( 'streamixer_category_slug', 'media-category' );
		$tag_slug = apply_filters( 'streamixer_tag_slug', 'media-tag' );

		// 註冊 CPT
		register_post_type( 'streamixer', array(
			'labels'       => array(
				'name'               => '素材組合',
				'singular_name'      => '素材組合',
				'add_new'            => '新增素材',
				'add_new_item'       => '新增素材組合',
				'edit_item'          => '編輯素材組合',
				'view_item'          => '檢視素材組合',
				'search_items'       => '搜尋素材',
				'not_found'          => '找不到素材',
				'not_found_in_trash' => '回收桶中無素材',
				'all_items'          => '所有素材',
				'menu_name'          => 'Streamixer',
			),
			'public'       => true,
			'has_archive'  => true,
			'rewrite'      => array( 'slug' => $cpt_slug ),
			'supports'     => array( 'title', 'editor', 'thumbnail' ),
			'menu_icon'    => 'dashicons-format-audio',
			'show_in_rest' => false,
		) );

		// 註冊分類法
		register_taxonomy( 'streamixer_category', 'streamixer', array(
			'labels'       => array(
				'name'          => '素材分類',
				'singular_name' => '分類',
				'add_new_item'  => '新增分類',
				'search_items'  => '搜尋分類',
			),
			'hierarchical' => true,
			'public'       => true,
			'rewrite'      => array( 'slug' => $cat_slug ),
			'show_in_rest' => true,
		) );

		register_taxonomy( 'streamixer_tag', 'streamixer', array(
			'labels'       => array(
				'name'          => '素材標籤',
				'singular_name' => '標籤',
				'add_new_item'  => '新增標籤',
				'search_items'  => '搜尋標籤',
			),
			'hierarchical' => false,
			'public'       => true,
			'rewrite'      => array( 'slug' => $tag_slug ),
			'show_in_rest' => true,
		) );

		// Meta box
		add_action( 'add_meta_boxes', array( __CLASS__, 'add_meta_boxes' ) );
		add_action( 'save_post_streamixer', array( __CLASS__, 'save_meta' ), 10, 2 );
		add_action( 'admin_enqueue_scripts', array( __CLASS__, 'enqueue_admin_scripts' ) );
	}

	public static function enqueue_admin_scripts( $hook ) {
		global $post_type;
		if ( 'streamixer' === $post_type && in_array( $hook, array( 'post.php', 'post-new.php' ), true ) ) {
			wp_enqueue_media();
		}
	}

	public static function add_meta_boxes() {
		add_meta_box(
			'streamixer_media',
			'素材檔案',
			array( __CLASS__, 'render_meta_box' ),
			'streamixer',
			'normal',
			'high'
		);
	}

	public static function render_meta_box( $post ) {
		wp_nonce_field( 'streamixer_save_meta', 'streamixer_nonce' );

		$audio_id      = get_post_meta( $post->ID, '_streamixer_audio_id', true );
		$background_id = get_post_meta( $post->ID, '_streamixer_background_id', true );
		$subtitle_id   = get_post_meta( $post->ID, '_streamixer_subtitle_id', true );
		$transcript_id = get_post_meta( $post->ID, '_streamixer_transcript_id', true );
		$sync_status   = get_post_meta( $post->ID, '_streamixer_sync_status', true );
		$sync_error    = get_post_meta( $post->ID, '_streamixer_sync_error', true );

		$audio_url      = $audio_id ? wp_get_attachment_url( $audio_id ) : '';
		$bg_url         = $background_id ? wp_get_attachment_url( $background_id ) : '';
		$sub_url        = $subtitle_id ? wp_get_attachment_url( $subtitle_id ) : '';
		$transcript_url = $transcript_id ? wp_get_attachment_url( $transcript_id ) : '';
		?>
		<style>
			.streamixer-field { margin-bottom: 15px; }
			.streamixer-field label { display: block; font-weight: bold; margin-bottom: 5px; }
			.streamixer-field .button { margin-right: 5px; }
			.streamixer-preview { color: #666; font-size: 12px; margin-top: 3px; }
			.streamixer-sync-status { padding: 8px 12px; border-radius: 4px; margin-top: 10px; }
			.streamixer-sync-status.synced { background: #d4edda; color: #155724; }
			.streamixer-sync-status.pending { background: #fff3cd; color: #856404; }
			.streamixer-sync-status.error { background: #f8d7da; color: #721c24; }
			.streamixer-shortcode-box { background: #f0f6fc; border: 1px solid #c3d4e6; border-radius: 4px; padding: 10px 14px; margin-bottom: 15px; display: flex; align-items: center; gap: 10px; }
			.streamixer-shortcode-box code { background: #fff; padding: 4px 10px; border-radius: 3px; font-size: 13px; flex: 1; user-select: all; }
			.streamixer-shortcode-box .button { flex-shrink: 0; }
			.streamixer-copied { color: #155724; font-size: 12px; display: none; }
		</style>

		<?php if ( 'publish' === $post->post_status ) : ?>
		<div class="streamixer-shortcode-box">
			<span>嵌入碼：</span>
			<code id="streamixer_shortcode">[streamixer id="<?php echo esc_attr( $post->ID ); ?>"]</code>
			<button type="button" class="button" onclick="navigator.clipboard.writeText(document.getElementById('streamixer_shortcode').textContent).then(function(){var el=document.getElementById('streamixer_copied');el.style.display='inline';setTimeout(function(){el.style.display='none'},2000)})">複製</button>
			<span class="streamixer-copied" id="streamixer_copied">✓ 已複製</span>
		</div>
		<?php endif; ?>

		<?php
		$files_cleaned = get_post_meta( $post->ID, '_streamixer_files_cleaned', true );

		$check_file = function( $attachment_id ) {
			if ( ! $attachment_id ) return null;
			$path = get_attached_file( $attachment_id );
			return ( $path && file_exists( $path ) );
		};

		$audio_exists      = $check_file( $audio_id );
		$bg_exists         = $check_file( $background_id );
		$sub_exists        = $check_file( $subtitle_id );
		$transcript_exists = $check_file( $transcript_id );

		$audio_display = $audio_url ? esc_html( basename( $audio_url ) ) : '未選擇';
		if ( ! $audio_id && $files_cleaned ) {
			$saved_name    = get_post_meta( $post->ID, '_streamixer_audio_id_filename', true );
			$audio_display = $saved_name ? esc_html( $saved_name ) . '（已同步至 Streamixer，本地已清除）' : '已同步至 Streamixer';
		} elseif ( $audio_id && ! $audio_exists ) {
			$audio_display = '⚠ 已選擇但檔案不存在（請重新上傳）';
		}

		$bg_display = $bg_url ? esc_html( basename( $bg_url ) ) : '未選擇';
		if ( ! $background_id && $files_cleaned ) {
			$saved_name = get_post_meta( $post->ID, '_streamixer_background_id_filename', true );
			$bg_display = $saved_name ? esc_html( $saved_name ) . '（已同步至 Streamixer，本地已清除）' : '已同步至 Streamixer';
		} elseif ( $background_id && ! $bg_exists ) {
			$bg_display = '⚠ 已選擇但檔案不存在（請重新上傳）';
		}

		$sub_display = $sub_url ? esc_html( basename( $sub_url ) ) : '未選擇';
		if ( ! $subtitle_id && $files_cleaned ) {
			$saved_name  = get_post_meta( $post->ID, '_streamixer_subtitle_id_filename', true );
			$sub_display = $saved_name ? esc_html( $saved_name ) . '（已同步至 Streamixer，本地已清除）' : '未選擇';
		} elseif ( $subtitle_id && ! $sub_exists ) {
			$sub_display = '⚠ 已選擇但檔案不存在（請重新上傳）';
		}

		$transcript_display = $transcript_url ? esc_html( basename( $transcript_url ) ) : '未選擇';
		if ( ! $transcript_id && $files_cleaned ) {
			$saved_name         = get_post_meta( $post->ID, '_streamixer_transcript_id_filename', true );
			$transcript_display = $saved_name ? esc_html( $saved_name ) . '（已同步至 Streamixer，本地已清除）' : '未選擇';
		} elseif ( ! $transcript_id ) {
			$saved_name = get_post_meta( $post->ID, '_streamixer_transcript_id_filename', true );
			if ( $saved_name ) {
				$transcript_display = esc_html( $saved_name ) . '（已同步至 Streamixer，本地已清除）';
			}
		} elseif ( $transcript_id && ! $transcript_exists ) {
			$transcript_display = '⚠ 已選擇但檔案不存在（請重新上傳）';
		}
		?>

		<div class="streamixer-field">
			<label>音檔（MP3 / WAV）*</label>
			<input type="hidden" name="streamixer_audio_id" id="streamixer_audio_id" value="<?php echo esc_attr( $audio_id ); ?>">
			<button type="button" class="button" id="streamixer_audio_btn">選擇音檔</button>
			<button type="button" class="button" id="streamixer_audio_clear">清除</button>
			<div class="streamixer-preview" id="streamixer_audio_preview"><?php echo $audio_display; ?></div>
		</div>

		<div class="streamixer-field">
			<label>背景圖片（JPG / PNG）*</label>
			<input type="hidden" name="streamixer_background_id" id="streamixer_background_id" value="<?php echo esc_attr( $background_id ); ?>">
			<button type="button" class="button" id="streamixer_background_btn">選擇圖片</button>
			<button type="button" class="button" id="streamixer_background_clear">清除</button>
			<div class="streamixer-preview" id="streamixer_background_preview"><?php echo $bg_display; ?></div>
		</div>

		<div class="streamixer-field">
			<label>字幕檔（SRT / VTT，選填）</label>
			<input type="hidden" name="streamixer_subtitle_id" id="streamixer_subtitle_id" value="<?php echo esc_attr( $subtitle_id ); ?>">
			<button type="button" class="button" id="streamixer_subtitle_btn">選擇字幕</button>
			<button type="button" class="button" id="streamixer_subtitle_clear">清除</button>
			<div class="streamixer-preview" id="streamixer_subtitle_preview"><?php echo $sub_display; ?></div>
		</div>

		<div class="streamixer-field">
			<label>逐字稿（TXT / PDF / DOC / DOCX / MD，選填）</label>
			<input type="hidden" name="streamixer_transcript_id" id="streamixer_transcript_id" value="<?php echo esc_attr( $transcript_id ); ?>">
			<button type="button" class="button" id="streamixer_transcript_btn">選擇逐字稿</button>
			<button type="button" class="button" id="streamixer_transcript_clear">清除</button>
			<div class="streamixer-preview" id="streamixer_transcript_preview"><?php echo $transcript_display; ?></div>
		</div>

		<?php
		$font_family = get_post_meta( $post->ID, '_streamixer_font', true );
		$fonts_data  = Streamixer_Fonts::fetch_all();
		?>
		<div class="streamixer-field">
			<label>字體</label>
			<select name="streamixer_font">
				<option value="">使用全站預設<?php echo $fonts_data['default_family'] ? '（' . esc_html( $fonts_data['default_family'] ) . '）' : ''; ?></option>
				<?php foreach ( $fonts_data['fonts'] as $f ) : ?>
					<option value="<?php echo esc_attr( $f['family_name'] ); ?>" <?php selected( $font_family, $f['family_name'] ); ?>>
						<?php echo esc_html( $f['family_name'] ); ?>
					</option>
				<?php endforeach; ?>
			</select>
			<p class="streamixer-preview">留空或選「使用全站預設」時，該素材字幕會套用 Streamixer 設定頁的全站預設字體。</p>
		</div>

		<?php if ( $sync_status ) : ?>
		<div class="streamixer-sync-status <?php echo esc_attr( $sync_status ); ?>">
			<?php
			switch ( $sync_status ) {
				case 'synced':
					$files_cleaned = get_post_meta( $post->ID, '_streamixer_files_cleaned', true );
					if ( $files_cleaned ) {
						echo '✓ 已同步至 Streamixer（本地檔案已清除，節省儲存空間）';
					} else {
						echo '✓ 已同步至 Streamixer';
					}
					break;
				case 'pending':
					echo '⏳ 等待同步...';
					break;
				case 'error':
					echo '✗ 同步失敗：' . esc_html( $sync_error );
					break;
			}
			?>
		</div>
		<?php endif; ?>

		<?php if ( 'synced' === $sync_status ) : ?>
		<div style="margin-top: 10px;">
			<a href="<?php echo esc_attr( Streamixer_API::get_download_url( $post->ID ) ); ?>" class="button" target="_blank">⬇ 匯出影片（MP4）</a>
		</div>
		<?php endif; ?>

		<script>
		jQuery(document).ready(function($) {
			function setupMediaButton(buttonId, clearId, inputId, previewId, type) {
				$('#' + buttonId).on('click', function(e) {
					e.preventDefault();
					var frame = wp.media({ multiple: false });
					frame.on('select', function() {
						var attachment = frame.state().get('selection').first().toJSON();
						$('#' + inputId).val(attachment.id);
						$('#' + previewId).text(attachment.filename);
					});
					frame.open();
				});
				$('#' + clearId).on('click', function(e) {
					e.preventDefault();
					$('#' + inputId).val('');
					$('#' + previewId).text('未選擇');
				});
			}
			setupMediaButton('streamixer_audio_btn', 'streamixer_audio_clear', 'streamixer_audio_id', 'streamixer_audio_preview', 'audio');
			setupMediaButton('streamixer_background_btn', 'streamixer_background_clear', 'streamixer_background_id', 'streamixer_background_preview', 'image');
			setupMediaButton('streamixer_subtitle_btn', 'streamixer_subtitle_clear', 'streamixer_subtitle_id', 'streamixer_subtitle_preview', 'text');
			setupMediaButton('streamixer_transcript_btn', 'streamixer_transcript_clear', 'streamixer_transcript_id', 'streamixer_transcript_preview', 'application');
		});
		</script>
		<?php
	}

	public static function save_meta( $post_id, $post ) {
		if ( ! isset( $_POST['streamixer_nonce'] ) || ! wp_verify_nonce( $_POST['streamixer_nonce'], 'streamixer_save_meta' ) ) {
			return;
		}
		if ( defined( 'DOING_AUTOSAVE' ) && DOING_AUTOSAVE ) {
			return;
		}
		if ( ! current_user_can( 'edit_post', $post_id ) ) {
			return;
		}

		$fields = array( 'streamixer_audio_id', 'streamixer_background_id', 'streamixer_subtitle_id', 'streamixer_transcript_id' );
		foreach ( $fields as $field ) {
			$value = isset( $_POST[ $field ] ) ? intval( $_POST[ $field ] ) : 0;
			update_post_meta( $post_id, '_' . $field, $value );
		}

		// 字體（字串）
		if ( isset( $_POST['streamixer_font'] ) ) {
			$font_value = sanitize_text_field( $_POST['streamixer_font'] );
			update_post_meta( $post_id, '_streamixer_font', $font_value );
		}

		// 同步至 Streamixer
		Streamixer_API::sync_composition( $post_id );
	}
}

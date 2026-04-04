<?php
/**
 * Shortcode [streamixer id="..."]
 */
class Streamixer_Shortcode {

	public static function init() {
		add_shortcode( 'streamixer', array( __CLASS__, 'render' ) );
	}

	public static function render( $atts ) {
		$atts = shortcode_atts( array(
			'id' => '',
		), $atts, 'streamixer' );

		if ( empty( $atts['id'] ) ) {
			return '<p class="streamixer-error-msg">Streamixer：缺少 id 屬性。</p>';
		}

		// 嘗試以 slug 查找
		$post = get_page_by_path( $atts['id'], OBJECT, 'streamixer' );

		// 嘗試以 post ID 查找
		if ( ! $post && is_numeric( $atts['id'] ) ) {
			$post = get_post( intval( $atts['id'] ) );
			if ( $post && 'streamixer' !== $post->post_type ) {
				$post = null;
			}
		}

		if ( ! $post ) {
			return '<p class="streamixer-error-msg">Streamixer：找不到指定的素材「' . esc_html( $atts['id'] ) . '」。</p>';
		}

		return Streamixer_Frontend::render_player( $post->ID );
	}
}

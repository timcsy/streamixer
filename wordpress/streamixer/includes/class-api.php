<?php
/**
 * Streamixer HTTP API 通訊
 */
class Streamixer_API {

	/**
	 * 取得 Streamixer 服務 URL（後端 PHP 使用，Docker 內部通訊）
	 */
	public static function get_service_url() {
		$url = get_option( 'streamixer_service_url', 'http://localhost:8080' );
		return rtrim( $url, '/' );
	}

	/**
	 * 取得前端播放 URL（瀏覽器使用）
	 */
	public static function get_public_url() {
		$url = get_option( 'streamixer_public_url', '' );
		if ( empty( $url ) ) {
			return self::get_service_url();
		}
		return rtrim( $url, '/' );
	}

	/**
	 * 同步素材至 Streamixer 服務
	 */
	public static function sync_composition( $post_id ) {
		$audio_id      = get_post_meta( $post_id, '_streamixer_audio_id', true );
		$background_id = get_post_meta( $post_id, '_streamixer_background_id', true );
		$subtitle_id   = get_post_meta( $post_id, '_streamixer_subtitle_id', true );

		if ( ! $audio_id || ! $background_id ) {
			// 音檔和背景為必填，如果有預設背景則使用
			if ( ! $background_id ) {
				$background_id = get_option( 'streamixer_default_background', 0 );
			}
			if ( ! $audio_id || ! $background_id ) {
				update_post_meta( $post_id, '_streamixer_sync_status', 'pending' );
				update_post_meta( $post_id, '_streamixer_sync_error', '缺少音檔或背景圖片' );
				return false;
			}
		}

		$composition_id = self::get_composition_id( $post_id );
		$url            = self::get_service_url() . '/upload/' . urlencode( $composition_id );

		// 準備 multipart 上傳
		$boundary = wp_generate_password( 24, false );
		$body     = '';

		// 音檔
		$audio_path = get_attached_file( $audio_id );
		if ( $audio_path && file_exists( $audio_path ) ) {
			$body .= self::build_multipart_field( $boundary, 'audio', $audio_path );
		}

		// 背景圖片
		$bg_path = get_attached_file( $background_id );
		if ( $bg_path && file_exists( $bg_path ) ) {
			$body .= self::build_multipart_field( $boundary, 'background', $bg_path );
		}

		// 字幕（選填）
		if ( $subtitle_id ) {
			$sub_path = get_attached_file( $subtitle_id );
			if ( $sub_path && file_exists( $sub_path ) ) {
				$body .= self::build_multipart_field( $boundary, 'subtitle', $sub_path );
			}
		}

		$body .= "--{$boundary}--\r\n";

		$response = wp_remote_post( $url, array(
			'timeout' => 120,
			'headers' => array(
				'Content-Type' => 'multipart/form-data; boundary=' . $boundary,
			),
			'body'    => $body,
		) );

		if ( is_wp_error( $response ) ) {
			update_post_meta( $post_id, '_streamixer_sync_status', 'error' );
			update_post_meta( $post_id, '_streamixer_sync_error', $response->get_error_message() );
			return false;
		}

		$code = wp_remote_retrieve_response_code( $response );
		if ( $code >= 200 && $code < 300 ) {
			update_post_meta( $post_id, '_streamixer_composition_id', $composition_id );
			update_post_meta( $post_id, '_streamixer_sync_status', 'synced' );
			update_post_meta( $post_id, '_streamixer_sync_error', '' );
			return true;
		}

		$body_response = wp_remote_retrieve_body( $response );
		update_post_meta( $post_id, '_streamixer_sync_status', 'error' );
		update_post_meta( $post_id, '_streamixer_sync_error', "HTTP {$code}: {$body_response}" );
		return false;
	}

	/**
	 * 取得素材組合的 Streamixer ID（使用 post slug）
	 */
	public static function get_composition_id( $post_id ) {
		$post = get_post( $post_id );
		return $post ? $post->post_name : 'post-' . $post_id;
	}

	/**
	 * 取得素材的 HLS 串流 URL
	 */
	public static function get_stream_url( $post_id ) {
		$composition_id = self::get_composition_id( $post_id );
		return self::get_public_url() . '/stream/' . rawurlencode( $composition_id ) . '/index.m3u8';
	}

	/**
	 * 建構 multipart form field
	 */
	private static function build_multipart_field( $boundary, $name, $file_path ) {
		$filename = basename( $file_path );
		$content  = file_get_contents( $file_path );
		$mime     = mime_content_type( $file_path ) ?: 'application/octet-stream';

		return "--{$boundary}\r\n" .
			"Content-Disposition: form-data; name=\"{$name}\"; filename=\"{$filename}\"\r\n" .
			"Content-Type: {$mime}\r\n\r\n" .
			$content . "\r\n";
	}
}

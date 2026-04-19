<?php
/**
 * 封裝對 Streamixer 後端 /fonts 端點的呼叫
 */
class Streamixer_Fonts {

	public static function fetch_all() {
		$url      = Streamixer_API::get_service_url() . '/fonts';
		$response = wp_remote_get( $url, array( 'timeout' => 10 ) );
		if ( is_wp_error( $response ) ) {
			return array( 'default_family' => '', 'fonts' => array(), 'error' => $response->get_error_message() );
		}
		$body = wp_remote_retrieve_body( $response );
		$data = json_decode( $body, true );
		if ( ! is_array( $data ) ) {
			return array( 'default_family' => '', 'fonts' => array(), 'error' => 'Invalid response' );
		}
		return array(
			'default_family' => isset( $data['default_family'] ) ? $data['default_family'] : '',
			'fonts'          => isset( $data['fonts'] ) && is_array( $data['fonts'] ) ? $data['fonts'] : array(),
		);
	}

	public static function upload( $file_path, $filename ) {
		$url      = Streamixer_API::get_service_url() . '/fonts';
		$boundary = wp_generate_password( 24, false );
		$body     = self::multipart_field( $boundary, 'font', $file_path, $filename );
		$body    .= "--{$boundary}--\r\n";

		$headers = array( 'Content-Type' => 'multipart/form-data; boundary=' . $boundary );
		$api_key = get_option( 'streamixer_api_key', '' );
		if ( ! empty( $api_key ) ) {
			$headers['X-API-Key'] = $api_key;
		}

		$response = wp_remote_post( $url, array(
			'timeout' => 60,
			'headers' => $headers,
			'body'    => $body,
		) );
		if ( is_wp_error( $response ) ) {
			return array( 'success' => false, 'message' => $response->get_error_message() );
		}
		$code = wp_remote_retrieve_response_code( $response );
		$body = wp_remote_retrieve_body( $response );
		if ( $code >= 200 && $code < 300 ) {
			return array( 'success' => true, 'data' => json_decode( $body, true ) );
		}
		$err = json_decode( $body, true );
		return array( 'success' => false, 'message' => isset( $err['error'] ) ? $err['error'] : $body );
	}

	public static function delete( $id ) {
		$url     = Streamixer_API::get_service_url() . '/fonts/' . rawurlencode( $id );
		$headers = array();
		$api_key = get_option( 'streamixer_api_key', '' );
		if ( ! empty( $api_key ) ) {
			$headers['X-API-Key'] = $api_key;
		}
		$response = wp_remote_request( $url, array(
			'method'  => 'DELETE',
			'timeout' => 10,
			'headers' => $headers,
		) );
		if ( is_wp_error( $response ) ) {
			return array( 'success' => false, 'message' => $response->get_error_message() );
		}
		$code = wp_remote_retrieve_response_code( $response );
		if ( 204 === $code ) {
			return array( 'success' => true );
		}
		$body = wp_remote_retrieve_body( $response );
		$err  = json_decode( $body, true );
		return array( 'success' => false, 'message' => isset( $err['error'] ) ? $err['error'] : $body );
	}

	public static function set_default( $family_name ) {
		$url     = Streamixer_API::get_service_url() . '/fonts/default';
		$headers = array( 'Content-Type' => 'application/json' );
		$api_key = get_option( 'streamixer_api_key', '' );
		if ( ! empty( $api_key ) ) {
			$headers['X-API-Key'] = $api_key;
		}
		$response = wp_remote_request( $url, array(
			'method'  => 'PUT',
			'timeout' => 10,
			'headers' => $headers,
			'body'    => wp_json_encode( array( 'family_name' => $family_name ) ),
		) );
		if ( is_wp_error( $response ) ) {
			return array( 'success' => false, 'message' => $response->get_error_message() );
		}
		$code = wp_remote_retrieve_response_code( $response );
		if ( $code >= 200 && $code < 300 ) {
			return array( 'success' => true );
		}
		$body = wp_remote_retrieve_body( $response );
		$err  = json_decode( $body, true );
		return array( 'success' => false, 'message' => isset( $err['error'] ) ? $err['error'] : $body );
	}

	/**
	 * 找出指定某 family name 的素材組合
	 */
	public static function compositions_using( $family_name ) {
		if ( empty( $family_name ) ) {
			return array();
		}
		$posts = get_posts( array(
			'post_type'      => 'streamixer',
			'post_status'    => array( 'publish', 'draft', 'pending' ),
			'posts_per_page' => -1,
			'meta_key'       => '_streamixer_font',
			'meta_value'     => $family_name,
		) );
		$out = array();
		foreach ( $posts as $p ) {
			$out[] = array( 'id' => $p->ID, 'title' => $p->post_title );
		}
		return $out;
	}

	private static function multipart_field( $boundary, $name, $file_path, $filename ) {
		$content = file_get_contents( $file_path );
		$mime    = mime_content_type( $file_path ) ?: 'application/octet-stream';
		return "--{$boundary}\r\n"
			. "Content-Disposition: form-data; name=\"{$name}\"; filename=\"{$filename}\"\r\n"
			. "Content-Type: {$mime}\r\n\r\n"
			. $content . "\r\n";
	}
}

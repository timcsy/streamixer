<?php get_header(); ?>

<div class="streamixer-single">
	<?php while ( have_posts() ) : the_post(); ?>
		<h1><?php the_title(); ?></h1>

		<div class="streamixer-single-meta">
			<span><?php echo get_the_date(); ?></span>
			<?php
			$categories = get_the_terms( get_the_ID(), 'streamixer_category' );
			if ( $categories && ! is_wp_error( $categories ) ) {
				$cat_names = wp_list_pluck( $categories, 'name' );
				echo ' · <span>' . esc_html( implode( ', ', $cat_names ) ) . '</span>';
			}
			$tags = get_the_terms( get_the_ID(), 'streamixer_tag' );
			if ( $tags && ! is_wp_error( $tags ) ) {
				$tag_names = wp_list_pluck( $tags, 'name' );
				echo ' · <span>' . esc_html( implode( ', ', $tag_names ) ) . '</span>';
			}
			?>
		</div>

		<?php echo Streamixer_Frontend::render_player( get_the_ID() ); ?>

		<?php if ( get_the_content() ) : ?>
		<div class="streamixer-single-content">
			<?php the_content(); ?>
		</div>
		<?php endif; ?>

	<?php endwhile; ?>
</div>

<?php get_footer(); ?>

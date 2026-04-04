<?php get_header(); ?>

<div class="streamixer-archive">
	<h1>素材組合</h1>

	<?php
	// 分類篩選
	$categories = get_terms( array(
		'taxonomy'   => 'streamixer_category',
		'hide_empty' => true,
	) );

	if ( $categories && ! is_wp_error( $categories ) ) :
		$current_cat = get_query_var( 'streamixer_category' );
		?>
		<div class="streamixer-filters">
			<a href="<?php echo get_post_type_archive_link( 'streamixer' ); ?>"
			   class="streamixer-filter-link <?php echo empty( $current_cat ) ? 'active' : ''; ?>">
				全部
			</a>
			<?php foreach ( $categories as $cat ) : ?>
				<a href="<?php echo get_term_link( $cat ); ?>"
				   class="streamixer-filter-link <?php echo ( $current_cat === $cat->slug ) ? 'active' : ''; ?>">
					<?php echo esc_html( $cat->name ); ?>
				</a>
			<?php endforeach; ?>
		</div>
	<?php endif; ?>

	<?php if ( have_posts() ) : ?>
		<ul class="streamixer-list">
			<?php while ( have_posts() ) : the_post(); ?>
				<li class="streamixer-list-item">
					<a href="<?php the_permalink(); ?>"><?php the_title(); ?></a>
					<div class="streamixer-list-meta">
						<span><?php echo get_the_date(); ?></span>
						<?php
						$cats = get_the_terms( get_the_ID(), 'streamixer_category' );
						if ( $cats && ! is_wp_error( $cats ) ) {
							echo '<span>' . esc_html( wp_list_pluck( $cats, 'name' )[0] ) . '</span>';
						}
						?>
					</div>
				</li>
			<?php endwhile; ?>
		</ul>

		<div class="streamixer-pagination">
			<?php the_posts_pagination(); ?>
		</div>
	<?php else : ?>
		<p>目前沒有素材組合。</p>
	<?php endif; ?>
</div>

<?php get_footer(); ?>

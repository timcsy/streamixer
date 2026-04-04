( function( blocks, element, components, data, blockEditor ) {
	var el = element.createElement;
	var useState = element.useState;
	var useEffect = element.useEffect;
	var SelectControl = components.SelectControl;
	var Placeholder = components.Placeholder;
	var useBlockProps = blockEditor.useBlockProps;

	blocks.registerBlockType( 'streamixer/player', {
		title: 'Streamixer 播放器',
		icon: 'format-audio',
		category: 'media',
		attributes: {
			compositionId: { type: 'number', default: 0 },
			compositionTitle: { type: 'string', default: '' },
		},

		edit: function( props ) {
			var compositionId = props.attributes.compositionId;
			var compositionTitle = props.attributes.compositionTitle;
			var blockProps = useBlockProps();
			var options = useState( [] );
			var optionsList = options[0];
			var setOptions = options[1];
			var loading = useState( true );
			var isLoading = loading[0];
			var setLoading = loading[1];

			useEffect( function() {
				wp.apiFetch( { path: '/wp/v2/streamixer?per_page=100&status=publish' } ).then( function( posts ) {
					var opts = [ { value: 0, label: '-- 選擇素材組合 --' } ];
					posts.forEach( function( p ) {
						opts.push( { value: p.id, label: p.title.rendered } );
					} );
					setOptions( opts );
					setLoading( false );
				} ).catch( function() {
					setOptions( [ { value: 0, label: '無法載入素材列表' } ] );
					setLoading( false );
				} );
			}, [] );

			if ( compositionId && compositionTitle ) {
				return el( 'div', blockProps,
					el( 'div', {
						style: {
							padding: '20px',
							background: '#1a1a2e',
							borderRadius: '8px',
							color: '#fff',
							textAlign: 'center',
						}
					},
						el( 'p', { style: { margin: '0 0 8px', fontSize: '13px', opacity: 0.6 } }, '🎵 Streamixer 播放器' ),
						el( 'p', { style: { margin: '0', fontSize: '16px', fontWeight: 'bold' } }, compositionTitle ),
						el( 'button', {
							onClick: function() { props.setAttributes( { compositionId: 0, compositionTitle: '' } ); },
							style: {
								marginTop: '12px', padding: '4px 14px',
								background: '#3b82f6', color: '#fff',
								border: 'none', borderRadius: '4px', cursor: 'pointer',
							},
						}, '更換素材' )
					)
				);
			}

			return el( 'div', blockProps,
				el( Placeholder, {
					icon: 'format-audio',
					label: 'Streamixer 播放器',
					instructions: '選擇要播放的素材組合',
				},
					isLoading
						? el( 'p', null, '載入中...' )
						: el( SelectControl, {
							value: compositionId,
							options: optionsList,
							onChange: function( val ) {
								var id = parseInt( val );
								var title = '';
								optionsList.forEach( function( o ) {
									if ( o.value === id ) title = o.label;
								} );
								props.setAttributes( { compositionId: id, compositionTitle: title } );
							},
						} )
				)
			);
		},

		save: function() {
			return null; // 動態渲染
		},
	} );
} )(
	window.wp.blocks,
	window.wp.element,
	window.wp.components,
	window.wp.data,
	window.wp.blockEditor
);

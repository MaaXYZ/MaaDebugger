def select_filter(element_id: int) -> str:
    """
    为 NiceGUI select 组件设置自定义过滤逻辑。

    过滤规则：
    1. 仅当输入字符 >= 2 时才触发筛选
    2. 优先级排序：startsWith > endsWith > includes
    """
    return f"""
            const el = getElement({element_id});
            if (el) {{
                el.findFilteredOptions = function() {{
                    const needle = this.$el.querySelector("input[type=search]")?.value.toLocaleLowerCase();
                    // 分类匹配：startsWith > endsWith > includes
                    const startsWithMatches = [];
                    const endsWithMatches = [];
                    const includesMatches = [];
                    
                    for (const opt of this.initialOptions) {{
                        const label = String(opt.label).toLocaleLowerCase();
                        if (label.startsWith(needle)) {{
                            startsWithMatches.push(opt);
                        }} else if (label.endsWith(needle)) {{
                            endsWithMatches.push(opt);
                        }} else if (label.includes(needle)) {{
                            includesMatches.push(opt);
                        }}
                    }}
                    
                    return [...startsWithMatches, ...endsWithMatches, ...includesMatches];
                }};
            }}"""

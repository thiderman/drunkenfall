package main

import "fmt"

var version, buildtime string

const kekkonen = `
                               :;ttjffLLGLLLLffjt;,
                           :;fGWWGWWWWWWWWWWWWGGGGGGLj,.
                        ;tfGGGGGGGGGWWWWGGGGGGGGGGGLLLLGj;
                      ifGGGGGGGGGGGGGGGGGGGGGGGLLLLLLLLLLLLj.
                    ;fGGGGGGGGGGGGGGGGGGGGGLLLLLLffffffffLffLt,
                  .iLLLGGGGGGGGGGGGGGGGGGLLLLLfffffffffjfffffjfi:
                 .fLLLLLGGGGGGGGGGGGGGGGLLLLLLLfffffjjjjjjjjjttjt
                .iffLfLGLGGGGGGGGGGGGGGGGGLLLLfffffffjjjjjjjtttii.
               .ijfffLLLLLLGGGGGGGGGGGGGGGLLLffLffffffjjjjjttti;;:
               ittfffffLLLLGGGGGGGGGGGGGGGLLLLLLLffffjjjjjttttti;:
              .iitjffffLLLLLGGGGGGGGGGGGGGGGGLLLffffjjttjjtitttii;:
              :ittffLLLLGGGLGGGGGGGGGWGGGGGGGGLLLLffjjttjjtittiiii:
             .,tjjffLLLGGGGGGGGGGGGWWWWWGGGGGGGLLLffjjtjfftitttti;:
              ,jffjjffLLGGGGGGGGGWWWWGGGWWWWWWWGGLLjjtjffjtttttti;:
             .;LLLLLLLLGGGGGGGGGWWWWGGWGGLGGGGWGGLfjjjffjjjjtttii,:.:.
             .;fffjjjfLLGGGGGGGGWGGGLfttitLGGGGGGLfjjjjfjjjjjti;,:;fLfi.
               .:;;,:...:,;jLLLfi.....:,;;iii:,,::,jjjjfjti;.. .;GGjiiji
           . ..:,:::..  ....,;;.:,:. .....  .:;;,..;tt;,.   ..:jLtiLfLGi
         ::    :,,;iii;,:::  .::iffjtttjjffffjt;,.        ..,;ijffGGLLL;
         ::  .:;itttjjjjjjj:tGG,:ifGGLLLLLLfffffji:  ..::::,tiiLtifGGLi
          ....:,;itttjjjjfijDDWGt .iGLLLLfffffffft,.:,,::,jjiiiLGLLGGt.
           ,:.:,,,;iittjji;WDDWGG,..fGGGLffffjjjji,,;,:,tjjji;iLjjLGW,
           .,,:..:,;iiii:.iDDDWGLft;ifLLLLLffjjji::ii,jjfjjti;iiiLfjj
              .:..::::..ifLWDDWGLLLLj;,,;;;iiii;;ttt;;fjfjtiiiijGLjf;
                  .:;;;;ti;jfLLjt;;ijjLfjjfffffjttti;tjjjjtitiijLfj;
                 .;tffjjjti::::;tLGGGGWWGLGGGLLLfLLLLfjjjjtti;;jji,
                .;i;fjjffLLGGWGWWWWWGGGGWGLLLLLffLLffjttttiii;,...
                 ,;ijjjjfffLGGLLGGGGGGGGGGLfjLLfffffjjiii;;;,:..:
                 .iifjttttjfLGGLLLLLLLLffffjtfLffffjjt;;;,;,:..:i
                  ,iLGi,:,,,;;iiiiiittttiiitjLLffjjjti:::,;,..:ff
                  .,jLfjtjjfLLGGGGGGGGLLLLfjjfLfjjjti;..;;,:,itG,
                    ijjjjjjttjjfffLLGGGLLLLfjjfjjjji;:.,;;,;tfW,
                    :;itjjffffLLGGGGGGGGGGLLLfjfjjj;:.,iiitjGDL
                    .:;ijjfffLLGGWWWWWWGGGGLLffjjti:..;ttijjEG,
                       .itjjjfLLLGGGWWWWGGLjji;,,::..:jjitWEj.
                        :;itjjjjjLLLGGGWGLi;,,::::::;ttiGEKW
                           ::,,;;iiittjj;,,,,,,;;;ijjjGEKKf:
                             .: ........::,,;;;ttjjGDKKKEL
                             ,j. ......:::,,;ttjjWE#KKKE,
                             ,fj,  .  ..,,;;iijDK#####f:
                             tWLf.   ..:,,;tfEKKK###Ef
                             jK##EL; .. :iDKKKKKKK#f:
                             LK#Effj;;::fDKKKKKKK#D.
                             EEi.     .:EKKKKKKKKL:
                                     :DitD#KKKKKG
                            ;i: .:,,::::.iK###Ki.
                           .iWD  .:::::,,:fE#Dt
                           ,L#D:ijttiiii;, :j:
                           jK#  ...::,,.  .
`

func Kekkonen() {
	// Kekkonen
	fmt.Println(kekkonen)
	fmt.Printf("\n\n                DrunkenFall %s (%s)\n\n\n", version, buildtime)
}
